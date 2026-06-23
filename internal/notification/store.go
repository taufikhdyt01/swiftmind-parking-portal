package notification

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"swiftmind/pkg/db"
)

//go:embed schema.sql
var schemaSQL string

// Notification is a message addressed to a user. ViolationID, when present,
// lets the UI deep-link to the related violation.
type Notification struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Message     string    `json:"message"`
	ViolationID string    `json:"violation_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store is the data-access layer for notifications.
type Store struct{ db *pgxpool.Pool }

// NewStore wraps a connection pool.
func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

// Migrate creates the notifications table if needed.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, schemaSQL)
	return err
}

// Insert stores a notification for a recipient, optionally linked to a violation.
func (s *Store) Insert(ctx context.Context, recipient, kind, message, violationID string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO notifications (recipient_email, kind, message, violation_id)
		 VALUES ($1, $2, $3, $4)`,
		recipient, kind, message, db.Nullable(violationID))
	return err
}

// List returns a recipient's notifications, newest first (capped).
func (s *Store) List(ctx context.Context, recipient string) ([]Notification, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, kind, message, violation_id, created_at
		   FROM notifications WHERE recipient_email = $1
		  ORDER BY created_at DESC LIMIT 50`, recipient)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var (
			n           Notification
			violationID *string
		)
		if err := rows.Scan(&n.ID, &n.Kind, &n.Message, &violationID, &n.CreatedAt); err != nil {
			return nil, err
		}
		if violationID != nil {
			n.ViolationID = *violationID
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
