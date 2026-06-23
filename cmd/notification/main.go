// Command notification consumes domain events and stores per-user notifications.
package main

import (
	"context"
	"encoding/json"
	"os"

	"swiftmind/internal/notification"
	"swiftmind/pkg/broker"
	"swiftmind/pkg/config"
	"swiftmind/pkg/db"
	"swiftmind/pkg/events"
	"swiftmind/pkg/httpx"
	"swiftmind/pkg/logging"
)

func main() {
	logger := logging.New("notification")
	ctx := context.Background()

	pool, err := db.Connect(ctx, config.Get("DATABASE_URL",
		"postgres://swiftmind:swiftmind@localhost:5432/swiftmind?sslmode=disable"))
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := notification.NewStore(pool)
	if err := store.Migrate(ctx); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}
	svc := notification.NewService(store)

	b, err := broker.Connect(config.Get("RABBITMQ_URL", "amqp://swiftmind:swiftmind@localhost:5672/"))
	if err != nil {
		logger.Error("rabbitmq connect", "err", err)
		os.Exit(1)
	}
	defer b.Close()

	if err := b.Consume("notification.violation-created", events.RoutingViolationCreated, func(body []byte) error {
		var evt events.ViolationCreated
		if err := json.Unmarshal(body, &evt); err != nil {
			return err
		}
		return svc.HandleViolationCreated(context.Background(), evt)
	}); err != nil {
		logger.Error("subscribe violation.created", "err", err)
		os.Exit(1)
	}

	if err := b.Consume("notification.payment-completed", events.RoutingPaymentCompleted, func(body []byte) error {
		var evt events.PaymentCompleted
		if err := json.Unmarshal(body, &evt); err != nil {
			return err
		}
		return svc.HandlePaymentCompleted(context.Background(), evt)
	}); err != nil {
		logger.Error("subscribe payment.completed", "err", err)
		os.Exit(1)
	}

	handler := notification.NewHandler(svc).Routes()
	if err := httpx.RunServer(":"+config.Get("PORT", "8085"), handler, logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
