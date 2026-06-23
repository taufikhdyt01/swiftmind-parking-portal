// Command rules owns fine-rule versioning: it stores rule versions, exposes the
// active ruleset (cached in Redis), and publishes new versions.
package main

import (
	"context"
	"os"
	"time"

	"parkwatch/internal/rules"
	"parkwatch/pkg/config"
	"parkwatch/pkg/db"
	"parkwatch/pkg/httpx"
	"parkwatch/pkg/logging"
	"parkwatch/pkg/redisx"
)

func main() {
	logger := logging.New("rules")
	ctx := context.Background()

	pool, err := db.Connect(ctx, config.Get("DATABASE_URL",
		"postgres://parkwatch:parkwatch@localhost:5432/parkwatch?sslmode=disable"))
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Redis is used to cache the active ruleset; failure is non-fatal.
	rdb, err := redisx.Connect(ctx, config.Get("REDIS_URL", "redis://localhost:6379/0"))
	if err != nil {
		logger.Warn("redis unavailable, caching disabled", "err", err)
		rdb = nil
	}

	store := rules.NewStore(pool)
	if err := store.Migrate(ctx); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	cache := rules.NewCache(rdb, config.Duration("RULES_CACHE_TTL", time.Hour), logger)
	svc := rules.NewService(store, cache, logger)
	if err := svc.Seed(ctx); err != nil {
		logger.Error("seed", "err", err)
		os.Exit(1)
	}

	handler := rules.NewHandler(svc).Routes()
	if err := httpx.RunServer(":"+config.Get("PORT", "8082"), handler, logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
