// Command payment owns invoices and payments: it creates an invoice when a
// violation is issued (async), charges via a mocked provider, and emits
// payment.completed so the violation is marked paid.
package main

import (
	"context"
	"encoding/json"
	"os"

	"swiftmind/internal/payment"
	"swiftmind/pkg/broker"
	"swiftmind/pkg/config"
	"swiftmind/pkg/db"
	"swiftmind/pkg/events"
	"swiftmind/pkg/httpx"
	"swiftmind/pkg/logging"
	"swiftmind/pkg/redisx"
)

func main() {
	logger := logging.New("payment")
	ctx := context.Background()

	pool, err := db.Connect(ctx, config.Get("DATABASE_URL",
		"postgres://swiftmind:swiftmind@localhost:5432/swiftmind?sslmode=disable"))
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Redis backs payment idempotency; failure is non-fatal.
	rdb, err := redisx.Connect(ctx, config.Get("REDIS_URL", "redis://localhost:6379/0"))
	if err != nil {
		logger.Warn("redis unavailable, idempotency lock disabled", "err", err)
		rdb = nil
	}

	var b *broker.Broker
	if br, err := broker.Connect(config.Get("RABBITMQ_URL", "amqp://swiftmind:swiftmind@localhost:5672/")); err != nil {
		logger.Warn("rabbitmq unavailable, events disabled", "err", err)
	} else {
		b = br
		defer b.Close()
	}

	store := payment.NewStore(pool)
	if err := store.Migrate(ctx); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	svc := payment.NewService(store, rdb, b, logger)

	// Create an invoice for each new violation.
	if b != nil {
		if err := b.Consume("payment.violation-created", events.RoutingViolationCreated, func(body []byte) error {
			var evt events.ViolationCreated
			if err := json.Unmarshal(body, &evt); err != nil {
				return err
			}
			return svc.CreateInvoiceFromEvent(context.Background(), evt)
		}); err != nil {
			logger.Error("subscribe violation.created", "err", err)
			os.Exit(1)
		}
	}

	handler := payment.NewHandler(svc).Routes()
	if err := httpx.RunServer(":"+config.Get("PORT", "8084"), handler, logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
