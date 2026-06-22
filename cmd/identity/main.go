// Command identity is the authentication service: it owns the users table,
// verifies credentials, and issues JWT access tokens.
package main

import (
	"context"
	"os"
	"time"

	"swiftmind/internal/identity"
	"swiftmind/pkg/config"
	"swiftmind/pkg/db"
	"swiftmind/pkg/httpx"
	"swiftmind/pkg/jwt"
	"swiftmind/pkg/logging"
)

func main() {
	logger := logging.New("identity")
	ctx := context.Background()

	pool, err := db.Connect(ctx, config.Get("DATABASE_URL",
		"postgres://swiftmind:swiftmind@localhost:5432/swiftmind?sslmode=disable"))
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := identity.NewStore(pool)
	if err := store.Migrate(ctx); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	jm := jwt.NewManager(
		config.Get("JWT_SECRET", "dev-secret"),
		config.Get("JWT_ISSUER", "swiftmind"),
		config.Duration("ACCESS_TOKEN_TTL", 24*time.Hour),
	)

	svc := identity.NewService(store, jm)
	if err := svc.Seed(ctx); err != nil {
		logger.Error("seed", "err", err)
		os.Exit(1)
	}

	handler := identity.NewHandler(svc).Routes()
	if err := httpx.RunServer(":"+config.Get("PORT", "8081"), handler, logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
