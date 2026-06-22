package violation

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// Violation is a persisted violation with its immutable fine snapshot.
type Violation struct {
	ID             string    `json:"id"`
	Plate          string    `json:"plate"`
	ViolationType  string    `json:"violation_type"`
	Location       string    `json:"location"`
	OccurredAt     time.Time `json:"occurred_at"`
	PhotoObject    string    `json:"-"`
	PhotoURL       string    `json:"photo_url,omitempty"`
	OwnerEmail     string    `json:"owner_email"`
	IssuedByEmail  string    `json:"issued_by_email"`

	RuleVersionID    string  `json:"rule_version_id"`
	RuleVersion      int     `json:"rule_version"`
	BaseAmount       int64   `json:"base_amount"`
	TimeMultiplier   float64 `json:"time_multiplier"`
	RepeatMultiplier float64 `json:"repeat_multiplier"`
	PriorUnpaidCount int     `json:"prior_unpaid_count"`
	FinalAmount      int64   `json:"final_amount"`

	PaymentStatus string    `json:"payment_status"`
	CreatedAt     time.Time `json:"created_at"`
}

// Store is the data-access layer for violations and the plate registry.
type Store struct{ db *pgxpool.Pool }

// NewStore wraps a connection pool.
func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

// Migrate creates the tables if needed.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, schemaSQL)
	return err
}

// SeedPlate registers a plate to an owner, ignoring duplicates.
func (s *Store) SeedPlate(ctx context.Context, plate, ownerEmail string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO plates (plate, owner_email) VALUES ($1, $2) ON CONFLICT (plate) DO NOTHING`,
		plate, ownerEmail)
	return err
}

// OwnerOfPlate returns the owner email for a plate, or "" if unregistered.
func (s *Store) OwnerOfPlate(ctx context.Context, plate string) (string, error) {
	var owner string
	err := s.db.QueryRow(ctx, `SELECT owner_email FROM plates WHERE plate = $1`, plate).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return owner, err
}

// CountPriorUnpaid counts unpaid violations on a plate in [since, before).
func (s *Store) CountPriorUnpaid(ctx context.Context, plate string, since, before time.Time) (int, error) {
	var n int
	err := s.db.QueryRow(ctx,
		`SELECT count(*) FROM violations
		  WHERE plate = $1 AND payment_status = 'unpaid'
		    AND occurred_at >= $2 AND occurred_at < $3`,
		plate, since, before).Scan(&n)
	return n, err
}

// Insert persists a violation and returns it with generated id/timestamp.
func (s *Store) Insert(ctx context.Context, v *Violation) (*Violation, error) {
	err := s.db.QueryRow(ctx,
		`INSERT INTO violations
		   (plate, violation_type, location, occurred_at, photo_object, owner_email, issued_by_email,
		    rule_version_id, rule_version, base_amount, time_multiplier, repeat_multiplier,
		    prior_unpaid_count, final_amount)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 RETURNING id, payment_status, created_at`,
		v.Plate, v.ViolationType, v.Location, v.OccurredAt, nullable(v.PhotoObject), nullable(v.OwnerEmail),
		v.IssuedByEmail, v.RuleVersionID, v.RuleVersion, v.BaseAmount, v.TimeMultiplier,
		v.RepeatMultiplier, v.PriorUnpaidCount, v.FinalAmount,
	).Scan(&v.ID, &v.PaymentStatus, &v.CreatedAt)
	return v, err
}

// List returns violations newest first. If ownerEmail is non-empty, only that
// owner's violations are returned (member view); empty returns all (officer).
func (s *Store) List(ctx context.Context, ownerEmail string) ([]Violation, error) {
	query := `SELECT id, plate, violation_type, location, occurred_at, photo_object, owner_email,
	                 issued_by_email, rule_version_id, rule_version, base_amount, time_multiplier,
	                 repeat_multiplier, prior_unpaid_count, final_amount, payment_status, created_at
	            FROM violations`
	args := []any{}
	if ownerEmail != "" {
		query += ` WHERE owner_email = $1`
		args = append(args, ownerEmail)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Violation
	for rows.Next() {
		v, err := scanViolation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

// Get returns a single violation by id, or (nil, nil) if not found.
func (s *Store) Get(ctx context.Context, id string) (*Violation, error) {
	v, err := scanViolation(s.db.QueryRow(ctx,
		`SELECT id, plate, violation_type, location, occurred_at, photo_object, owner_email,
		        issued_by_email, rule_version_id, rule_version, base_amount, time_multiplier,
		        repeat_multiplier, prior_unpaid_count, final_amount, payment_status, created_at
		   FROM violations WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return v, err
}

// MarkPaid flips a violation to paid (driven by the payment.completed event).
func (s *Store) MarkPaid(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `UPDATE violations SET payment_status = 'paid' WHERE id = $1`, id)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanViolation(rs rowScanner) (*Violation, error) {
	var (
		v           Violation
		photo       *string
		owner       *string
	)
	if err := rs.Scan(
		&v.ID, &v.Plate, &v.ViolationType, &v.Location, &v.OccurredAt, &photo, &owner,
		&v.IssuedByEmail, &v.RuleVersionID, &v.RuleVersion, &v.BaseAmount, &v.TimeMultiplier,
		&v.RepeatMultiplier, &v.PriorUnpaidCount, &v.FinalAmount, &v.PaymentStatus, &v.CreatedAt,
	); err != nil {
		return nil, err
	}
	if photo != nil {
		v.PhotoObject = *photo
	}
	if owner != nil {
		v.OwnerEmail = *owner
	}
	return &v, nil
}

// nullable converts an empty string to nil for nullable columns.
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
