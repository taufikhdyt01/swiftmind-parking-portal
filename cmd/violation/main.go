// Command violation handles violation submission: it prices each violation
// against the active ruleset, stores an immutable snapshot, uploads the photo to
// MinIO, and emits violation.created.
package main

import (
	"context"
	"os"

	"swiftmind/internal/violation"
	"swiftmind/pkg/broker"
	"swiftmind/pkg/config"
	"swiftmind/pkg/db"
	"swiftmind/pkg/httpx"
	"swiftmind/pkg/logging"
	"swiftmind/pkg/objstore"
)

func main() {
	logger := logging.New("violation")
	ctx := context.Background()

	pool, err := db.Connect(ctx, config.Get("DATABASE_URL",
		"postgres://swiftmind:swiftmind@localhost:5432/swiftmind?sslmode=disable"))
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	photos, err := objstore.New(ctx, objstore.Config{
		Endpoint:  config.Get("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey: config.Get("MINIO_ROOT_USER", "swiftmind"),
		SecretKey: config.Get("MINIO_ROOT_PASSWORD", "swiftmind123"),
		UseSSL:    config.Bool("MINIO_USE_SSL", false),
		Bucket:    config.Get("MINIO_BUCKET", "violation-photos"),
	})
	if err != nil {
		logger.Error("minio connect", "err", err)
		os.Exit(1)
	}

	// RabbitMQ is used to announce new violations; failure is non-fatal.
	var b *broker.Broker
	if br, err := broker.Connect(config.Get("RABBITMQ_URL", "amqp://swiftmind:swiftmind@localhost:5672/")); err != nil {
		logger.Warn("rabbitmq unavailable, events disabled", "err", err)
	} else {
		b = br
		defer b.Close()
	}

	store := violation.NewStore(pool)
	if err := store.Migrate(ctx); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	rules := violation.NewRulesClient(config.Get("RULES_URL", "http://localhost:8082"))
	svc := violation.NewService(store, photos, rules, b, logger)
	if err := svc.Seed(ctx); err != nil {
		logger.Error("seed", "err", err)
		os.Exit(1)
	}

	handler := violation.NewHandler(svc).Routes()
	if err := httpx.RunServer(":"+config.Get("PORT", "8083"), handler, logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
