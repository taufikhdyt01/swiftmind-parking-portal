package rules

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"parkwatch/pkg/db"
	"parkwatch/pkg/fine"
)

//go:embed schema.sql
var schemaSQL string

// Version is one published fine-rule version.
type Version struct {
	ID        string       `json:"id"`
	Version   int          `json:"version"`
	IsActive  bool         `json:"is_active"`
	Ruleset   fine.Ruleset `json:"ruleset"`
	CreatedBy string       `json:"created_by"`
	CreatedAt time.Time    `json:"created_at"`
}

// Store is the data-access layer for rule_versions.
type Store struct{ db *pgxpool.Pool }

// NewStore wraps a connection pool.
func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

// Migrate creates the rule_versions table if needed.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, schemaSQL)
	return err
}

// Count returns the number of rule versions (used to decide whether to seed).
func (s *Store) Count(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT count(*) FROM rule_versions`).Scan(&n)
	return n, err
}

// Active returns the currently active version, or (nil, nil) if none exists.
func (s *Store) Active(ctx context.Context) (*Version, error) {
	return scanOne(s.db.QueryRow(ctx,
		`SELECT id, version, is_active, config, created_by, created_at
		   FROM rule_versions WHERE is_active`))
}

// List returns all versions, newest first.
func (s *Store) List(ctx context.Context) ([]Version, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, version, is_active, config, created_by, created_at
		   FROM rule_versions ORDER BY version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Version
	for rows.Next() {
		v, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

// Publish atomically deactivates the current version and inserts a new active
// one with the next version number. Past versions are left untouched.
func (s *Store) Publish(ctx context.Context, ruleset fine.Ruleset, createdBy string) (*Version, error) {
	config, err := json.Marshal(ruleset)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	var nextVersion int
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(max(version), 0) + 1 FROM rule_versions`).Scan(&nextVersion); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE rule_versions SET is_active = false WHERE is_active`); err != nil {
		return nil, err
	}

	v := &Version{Ruleset: ruleset}
	if err := tx.QueryRow(ctx,
		`INSERT INTO rule_versions (version, is_active, config, created_by)
		 VALUES ($1, true, $2, $3)
		 RETURNING id, version, is_active, created_by, created_at`,
		nextVersion, config, createdBy,
	).Scan(&v.ID, &v.Version, &v.IsActive, &v.CreatedBy, &v.CreatedAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

func scanRow(rs db.RowScanner) (*Version, error) {
	var (
		v      Version
		config []byte
	)
	if err := rs.Scan(&v.ID, &v.Version, &v.IsActive, &config, &v.CreatedBy, &v.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(config, &v.Ruleset); err != nil {
		return nil, err
	}
	return &v, nil
}

func scanOne(row pgx.Row) (*Version, error) {
	v, err := scanRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return v, err
}
