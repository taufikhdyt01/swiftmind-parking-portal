package notification

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"swiftmind/pkg/events"
)

// Notification kinds.
const (
	KindViolationIssued  = "violation_issued"
	KindPaymentCompleted = "payment_completed"
)

// Service stores notifications derived from domain events.
type Service struct{ store *Store }

// NewService wraps the store.
func NewService(store *Store) *Service { return &Service{store: store} }

// HandleViolationCreated notifies the plate owner that a fine was issued.
func (s *Service) HandleViolationCreated(ctx context.Context, evt events.ViolationCreated) error {
	if evt.OwnerEmail == "" {
		return nil // unregistered plate — nobody to notify
	}
	msg := fmt.Sprintf("A %s violation was issued for plate %s. Fine: %s.",
		humanize(evt.ViolationType), evt.Plate, idr(evt.FinalAmount))
	return s.store.Insert(ctx, evt.OwnerEmail, KindViolationIssued, msg)
}

// HandlePaymentCompleted notifies the member that their payment succeeded.
func (s *Service) HandlePaymentCompleted(ctx context.Context, evt events.PaymentCompleted) error {
	if evt.OwnerEmail == "" {
		return nil
	}
	msg := fmt.Sprintf("Your payment of %s was successful. Transaction %s.",
		idr(evt.Amount), evt.TransactionID)
	return s.store.Insert(ctx, evt.OwnerEmail, KindPaymentCompleted, msg)
}

// List returns a recipient's notifications.
func (s *Service) List(ctx context.Context, email string) ([]Notification, error) {
	return s.store.List(ctx, email)
}

// humanize turns "no_parking_zone" into "no parking zone".
func humanize(s string) string {
	return strings.ReplaceAll(s, "_", " ")
}

// idr formats an integer rupiah amount as "Rp 1.500.000".
func idr(amount int64) string {
	s := strconv.FormatInt(amount, 10)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return "Rp " + strings.Join(parts, ".")
}
