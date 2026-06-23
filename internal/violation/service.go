package violation

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/google/uuid"

	"parkwatch/pkg/broker"
	"parkwatch/pkg/domain"
	"parkwatch/pkg/events"
	"parkwatch/pkg/fine"
	"parkwatch/pkg/objstore"
)

// repeatWindow is the look-back period for prior unpaid violations.
const repeatWindow = 90 * 24 * time.Hour

// ErrInvalidInput marks a user-facing validation failure.
var ErrInvalidInput = errors.New("invalid input")

// ErrNotFound marks a missing violation or photo.
var ErrNotFound = errors.New("not found")

// Service holds the violation business logic and its collaborators.
type Service struct {
	store  *Store
	photos *objstore.Store
	rules  *RulesClient
	broker *broker.Broker
	logger *slog.Logger
}

// NewService wires the collaborators. broker may be nil (events disabled).
func NewService(store *Store, photos *objstore.Store, rules *RulesClient, b *broker.Broker, logger *slog.Logger) *Service {
	return &Service{store: store, photos: photos, rules: rules, broker: b, logger: logger}
}

// Photo carries an uploaded image to store.
type Photo struct {
	Reader      io.Reader
	Size        int64
	ContentType string
	Filename    string
}

// CreateInput is the officer-supplied data for a new violation.
type CreateInput struct {
	Plate         string
	ViolationType string
	Location      string
	OccurredAt    time.Time
	IssuedByEmail string
}

// Seed registers the demo member's plate.
func (s *Service) Seed(ctx context.Context) error {
	return s.store.SeedPlate(ctx, "B1234ABC", "member@parkwatch.test")
}

// Create prices a violation against the active ruleset, persists it with an
// immutable snapshot, stores the photo, and emits violation.created.
func (s *Service) Create(ctx context.Context, in CreateInput, photo *Photo) (*Violation, error) {
	if in.Plate == "" || in.Location == "" || in.OccurredAt.IsZero() {
		return nil, ErrInvalidInput
	}
	if !domain.ViolationType(in.ViolationType).Valid() {
		return nil, ErrInvalidInput
	}

	// 1. Fetch the active ruleset (synchronous service-to-service call).
	active, err := s.rules.Active(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Count prior unpaid violations on this plate in the look-back window.
	priorUnpaid, err := s.store.CountPriorUnpaid(ctx, in.Plate, in.OccurredAt.Add(-repeatWindow), in.OccurredAt)
	if err != nil {
		return nil, err
	}

	// 3. Calculate the fine.
	breakdown, err := active.Ruleset.Calculate(fine.Input{
		ViolationType:    in.ViolationType,
		OccurredAt:       in.OccurredAt,
		PriorUnpaidCount: priorUnpaid,
	})
	if err != nil {
		return nil, ErrInvalidInput
	}

	// 4. Store the photo (if provided).
	var photoKey string
	if photo != nil && photo.Size > 0 {
		photoKey = "violations/" + uuid.NewString() + path.Ext(photo.Filename)
		if err := s.photos.Put(ctx, photoKey, photo.ContentType, photo.Reader, photo.Size); err != nil {
			return nil, err
		}
	}

	// 5. Resolve plate ownership (denormalized onto the row).
	owner, err := s.store.OwnerOfPlate(ctx, in.Plate)
	if err != nil {
		return nil, err
	}

	// 6. Persist with the calculation snapshot.
	v := &Violation{
		Plate:            in.Plate,
		ViolationType:    in.ViolationType,
		Location:         in.Location,
		OccurredAt:       in.OccurredAt,
		PhotoObject:      photoKey,
		OwnerEmail:       owner,
		IssuedByEmail:    in.IssuedByEmail,
		RuleVersionID:    active.ID,
		RuleVersion:      active.Version,
		BaseAmount:       breakdown.BaseAmount,
		TimeMultiplier:   breakdown.TimeMultiplier,
		RepeatMultiplier: breakdown.RepeatMultiplier,
		PriorUnpaidCount: breakdown.PriorUnpaidCount,
		FinalAmount:      breakdown.FinalAmount,
	}
	if _, err := s.store.Insert(ctx, v); err != nil {
		return nil, err
	}

	// 7. Announce the violation so an invoice can be created asynchronously.
	s.publishCreated(ctx, v)

	attachPhotoURL(v)
	return v, nil
}

// Get returns a single violation visible to the caller — officers see any,
// members only their own — with its photo URL attached. A member requesting
// someone else's violation gets ErrNotFound (no existence leak).
func (s *Service) Get(ctx context.Context, id, role, email string) (*Violation, error) {
	v, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrNotFound
	}
	if role == domain.RoleMember.String() && v.OwnerEmail != email {
		return nil, ErrNotFound
	}
	attachPhotoURL(v)
	return v, nil
}

// MarkPaid flips a violation to paid. Driven by the payment.completed event so
// the member no longer owes the fine and it stops counting as a prior unpaid.
func (s *Service) MarkPaid(ctx context.Context, violationID string) error {
	return s.store.MarkPaid(ctx, violationID)
}

// Photo opens the stored photo for a violation, returning its content type.
func (s *Service) Photo(ctx context.Context, id string) (io.ReadCloser, string, error) {
	v, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if v == nil || v.PhotoObject == "" {
		return nil, "", ErrNotFound
	}
	return s.photos.Get(ctx, v.PhotoObject)
}

// List returns violations visible to the caller: officers see all, members see
// only violations on their registered plates.
func (s *Service) List(ctx context.Context, role, email string) ([]Violation, error) {
	owner := ""
	if role == domain.RoleMember.String() {
		owner = email
	}
	violations, err := s.store.List(ctx, owner)
	if err != nil {
		return nil, err
	}
	for i := range violations {
		attachPhotoURL(&violations[i])
	}
	return violations, nil
}

func (s *Service) publishCreated(ctx context.Context, v *Violation) {
	if s.broker == nil {
		return
	}
	evt := events.ViolationCreated{
		ViolationID:   v.ID,
		Plate:         v.Plate,
		ViolationType: v.ViolationType,
		OwnerEmail:    v.OwnerEmail,
		IssuedByEmail: v.IssuedByEmail,
		FinalAmount:   v.FinalAmount,
		CreatedAt:     v.CreatedAt,
	}
	if err := s.broker.Publish(ctx, events.RoutingViolationCreated, evt); err != nil {
		s.logger.Warn("publish violation.created failed", "err", err, "violation_id", v.ID)
	}
}

// attachPhotoURL sets a gateway-relative URL the browser can fetch (cookies
// ride along, the photo streams through the service).
func attachPhotoURL(v *Violation) {
	if v.PhotoObject != "" {
		v.PhotoURL = "/api/violations/" + v.ID + "/photo"
	}
}
