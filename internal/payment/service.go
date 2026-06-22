package payment

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"swiftmind/pkg/broker"
	"swiftmind/pkg/domain"
	"swiftmind/pkg/events"
)

var (
	// ErrNotFound is returned when an invoice does not exist.
	ErrNotFound = errors.New("invoice not found")
	// ErrForbidden is returned when a member tries to pay someone else's invoice.
	ErrForbidden = errors.New("forbidden")
	// ErrInvalidScenario is returned for an unknown payment scenario.
	ErrInvalidScenario = errors.New("invalid scenario")
	// ErrBusy is returned when a concurrent payment for the same invoice is in flight.
	ErrBusy = errors.New("payment already in progress")
)

// lockTTL bounds how long a payment idempotency lock is held.
const lockTTL = 15 * time.Second

// Service holds the payment business logic.
type Service struct {
	store  *Store
	redis  *redis.Client
	broker *broker.Broker
	logger *slog.Logger
}

// NewService wires the collaborators. redis and broker may be nil.
func NewService(store *Store, rdb *redis.Client, b *broker.Broker, logger *slog.Logger) *Service {
	return &Service{store: store, redis: rdb, broker: b, logger: logger}
}

// PayResult is returned to the caller after a charge attempt.
type PayResult struct {
	Status        string   `json:"status"`
	TransactionID string   `json:"transaction_id"`
	Invoice       *Invoice `json:"invoice"`
}

// CreateInvoiceFromEvent creates an invoice when a violation is issued.
func (s *Service) CreateInvoiceFromEvent(ctx context.Context, evt events.ViolationCreated) error {
	return s.store.CreateInvoice(ctx, &Invoice{
		ViolationID:   evt.ViolationID,
		Plate:         evt.Plate,
		ViolationType: evt.ViolationType,
		OwnerEmail:    evt.OwnerEmail,
		Amount:        evt.FinalAmount,
	})
}

// List returns invoices visible to the caller: officers see all, members see
// only their own.
func (s *Service) List(ctx context.Context, role, email string) ([]Invoice, error) {
	owner := ""
	if role == domain.RoleMember.String() {
		owner = email
	}
	return s.store.List(ctx, owner)
}

// Pay charges an invoice via the mocked provider. callerEmail must own the
// invoice. A Redis lock makes concurrent attempts safe; an already-paid invoice
// short-circuits idempotently.
func (s *Service) Pay(ctx context.Context, invoiceID, scenario, callerEmail string) (*PayResult, error) {
	if scenario != ScenarioSuccess && scenario != ScenarioFailed {
		return nil, ErrInvalidScenario
	}

	inv, err := s.store.Get(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrNotFound
	}
	if callerEmail != "" && inv.OwnerEmail != callerEmail {
		return nil, ErrForbidden
	}
	if inv.Status == "paid" {
		// Idempotent: already settled.
		return &PayResult{Status: StatusPaid, Invoice: inv}, nil
	}

	// Idempotency lock: prevent two concurrent charges for the same invoice.
	release, ok := s.acquireLock(ctx, invoiceID)
	if !ok {
		return nil, ErrBusy
	}
	defer release()

	result := Charge(inv.ID, inv.Amount, scenario)

	if err := s.store.InsertPayment(ctx, Payment{
		InvoiceID:     inv.ID,
		Amount:        inv.Amount,
		Scenario:      scenario,
		Status:        result.Status,
		TransactionID: result.TransactionID,
	}); err != nil {
		return nil, err
	}

	if result.Status == StatusPaid {
		if err := s.store.MarkPaid(ctx, inv.ID); err != nil {
			return nil, err
		}
		inv.Status = "paid"
		s.publishCompleted(ctx, inv, result.TransactionID)
	}

	return &PayResult{Status: result.Status, TransactionID: result.TransactionID, Invoice: inv}, nil
}

// acquireLock takes a short Redis lock for the invoice. With no Redis it is a
// no-op (the invoice status check still guards against double payment).
func (s *Service) acquireLock(ctx context.Context, invoiceID string) (func(), bool) {
	if s.redis == nil {
		return func() {}, true
	}
	key := "pay:lock:" + invoiceID
	ok, err := s.redis.SetNX(ctx, key, "1", lockTTL).Result()
	if err != nil {
		// On Redis error, fall back to allowing the charge (status check guards).
		return func() {}, true
	}
	if !ok {
		return nil, false
	}
	return func() { s.redis.Del(context.Background(), key) }, true
}

func (s *Service) publishCompleted(ctx context.Context, inv *Invoice, txnID string) {
	if s.broker == nil {
		return
	}
	evt := events.PaymentCompleted{
		InvoiceID:     inv.ID,
		ViolationID:   inv.ViolationID,
		OwnerEmail:    inv.OwnerEmail,
		Amount:        inv.Amount,
		TransactionID: txnID,
		PaidAt:        time.Now(),
	}
	if err := s.broker.Publish(ctx, events.RoutingPaymentCompleted, evt); err != nil {
		s.logger.Warn("publish payment.completed failed", "err", err, "invoice_id", inv.ID)
	}
}
