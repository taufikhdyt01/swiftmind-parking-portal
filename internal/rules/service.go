package rules

import (
	"context"
	"errors"
	"log/slog"

	"parkwatch/pkg/fine"
)

// ErrNoActiveRuleset means no rule version has been published yet.
var ErrNoActiveRuleset = errors.New("no active ruleset")

// Service holds the rules business logic.
type Service struct {
	store  *Store
	cache  *Cache
	logger *slog.Logger
}

// NewService wires the store and cache.
func NewService(store *Store, cache *Cache, logger *slog.Logger) *Service {
	return &Service{store: store, cache: cache, logger: logger}
}

// Seed publishes the assignment's day-one ruleset as version 1 on first start.
func (s *Service) Seed(ctx context.Context) error {
	n, err := s.store.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = s.store.Publish(ctx, fine.DefaultRuleset(), "system")
	return err
}

// Active returns the active version, reading through the Redis cache.
func (s *Service) Active(ctx context.Context) (*Version, error) {
	if v, ok := s.cache.GetActive(ctx); ok {
		return v, nil
	}
	v, err := s.store.Active(ctx)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrNoActiveRuleset
	}
	s.cache.SetActive(ctx, v)
	return v, nil
}

// List returns the full version history, newest first.
func (s *Service) List(ctx context.Context) ([]Version, error) {
	return s.store.List(ctx)
}

// Publish validates and stores a new active version, then invalidates the cache.
// Existing versions (and the fines snapshotted against them) are never modified.
func (s *Service) Publish(ctx context.Context, ruleset fine.Ruleset, createdBy string) (*Version, error) {
	if err := ruleset.Validate(); err != nil {
		return nil, err
	}
	v, err := s.store.Publish(ctx, ruleset, createdBy)
	if err != nil {
		return nil, err
	}
	s.cache.Invalidate(ctx)
	return v, nil
}
