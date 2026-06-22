// Command gateway is the single entrypoint between the frontend and the backend
// services. It verifies JWTs, enforces RBAC, and reverse-proxies to services.
package main

import (
	"os"
	"time"

	"swiftmind/internal/gateway"
	"swiftmind/pkg/config"
	"swiftmind/pkg/httpx"
	"swiftmind/pkg/jwt"
	"swiftmind/pkg/logging"
)

func main() {
	logger := logging.New("gateway")

	jm := jwt.NewManager(
		config.Get("JWT_SECRET", "dev-secret"),
		config.Get("JWT_ISSUER", "swiftmind"),
		config.Duration("ACCESS_TOKEN_TTL", 24*time.Hour),
	)

	gw := gateway.New(jm, gateway.Config{
		IdentityURL:  config.Get("IDENTITY_URL", "http://localhost:8081"),
		RulesURL:     config.Get("RULES_URL", "http://localhost:8082"),
		ViolationURL: config.Get("VIOLATION_URL", "http://localhost:8083"),
		PaymentURL:   config.Get("PAYMENT_URL", "http://localhost:8084"),
		CookieSecure: config.Bool("COOKIE_SECURE", false),
		AllowOrigin:  config.Get("CORS_ALLOW_ORIGIN", "http://localhost:3000"),
	})

	if err := httpx.RunServer(":"+config.Get("PORT", "8080"), gw.Router(), logger); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
