package payment

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"swiftmind/pkg/db"
)

//go:embed schema.sql
var schemaSQL string

// Invoice is an amount owed for a single violation. TransactionID is the
// reference of the successful payment, present once the invoice is paid.
type Invoice struct {
	ID            string    `json:"id"`
	ViolationID   string    `json:"violation_id"`
	Plate         string    `json:"plate"`
	ViolationType string    `json:"violation_type"`
	OwnerEmail    string    `json:"owner_email"`
	IssuedByEmail string    `json:"issued_by_email"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	TransactionID string    `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// invoiceSelect joins each invoice to its latest successful payment so the
// transaction reference can be shown once paid.
const invoiceSelect = `
	SELECT i.id, i.violation_id, i.plate, i.violation_type, i.owner_email, i.issued_by_email,
	       i.amount, i.status, i.created_at, p.transaction_id
	  FROM invoices i
	  LEFT JOIN LATERAL (
	      SELECT transaction_id FROM payments
	       WHERE invoice_id = i.id AND status = 'paid'
	       ORDER BY created_at DESC LIMIT 1
	  ) p ON true`

// Payment is a recorded charge attempt against an invoice.
type Payment struct {
	InvoiceID     string
	Amount        int64
	Scenario      string
	Status        string
	TransactionID string
}

// Store is the data-access layer for invoices and payments.
type Store struct{ db *pgxpool.Pool }

// NewStore wraps a connection pool.
func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

// Migrate creates the tables if needed.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, schemaSQL)
	return err
}

// CreateInvoice inserts an invoice, ignoring duplicates (idempotent on the
// violation, so a redelivered event does not create a second invoice).
func (s *Store) CreateInvoice(ctx context.Context, inv *Invoice) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO invoices (violation_id, plate, violation_type, owner_email, issued_by_email, amount)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (violation_id) DO NOTHING`,
		inv.ViolationID, inv.Plate, inv.ViolationType, db.Nullable(inv.OwnerEmail),
		db.Nullable(inv.IssuedByEmail), inv.Amount)
	return err
}

// Get returns an invoice by id, or (nil, nil) if not found.
func (s *Store) Get(ctx context.Context, id string) (*Invoice, error) {
	inv, err := scanInvoice(s.db.QueryRow(ctx, invoiceSelect+` WHERE i.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return inv, err
}

// List returns invoices newest first, optionally filtered to one owner.
func (s *Store) List(ctx context.Context, ownerEmail string) ([]Invoice, error) {
	query := invoiceSelect
	args := []any{}
	if ownerEmail != "" {
		query += ` WHERE i.owner_email = $1`
		args = append(args, ownerEmail)
	}
	query += ` ORDER BY i.created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *inv)
	}
	return out, rows.Err()
}

// MarkPaid flips an invoice to paid.
func (s *Store) MarkPaid(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `UPDATE invoices SET status = 'paid' WHERE id = $1`, id)
	return err
}

// InsertPayment records a charge attempt.
func (s *Store) InsertPayment(ctx context.Context, p Payment) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO payments (invoice_id, amount, scenario, status, transaction_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		p.InvoiceID, p.Amount, p.Scenario, p.Status, p.TransactionID)
	return err
}

func scanInvoice(rs db.RowScanner) (*Invoice, error) {
	var (
		inv      Invoice
		owner    *string
		issuedBy *string
		txn      *string
	)
	if err := rs.Scan(&inv.ID, &inv.ViolationID, &inv.Plate, &inv.ViolationType,
		&owner, &issuedBy, &inv.Amount, &inv.Status, &inv.CreatedAt, &txn); err != nil {
		return nil, err
	}
	if owner != nil {
		inv.OwnerEmail = *owner
	}
	if issuedBy != nil {
		inv.IssuedByEmail = *issuedBy
	}
	if txn != nil {
		inv.TransactionID = *txn
	}
	return &inv, nil
}
