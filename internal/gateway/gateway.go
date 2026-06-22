// Package gateway implements the single entrypoint between the frontend and the
// backend services. It performs JWT verification, role-based access control, and
// reverse-proxies authenticated requests to downstream services, forwarding the
// caller's identity as trusted X-User-* headers.
package gateway

import (
	"net/http"
	"time"

	"swiftmind/pkg/jwt"
)

// Config holds the downstream service URLs and cookie behaviour.
type Config struct {
	IdentityURL  string
	RulesURL     string
	ViolationURL string
	PaymentURL   string
	CookieName   string
	CookieSecure bool
	AllowOrigin  string
}

// Gateway wires the JWT manager, config, and an HTTP client for service calls.
type Gateway struct {
	jwt    *jwt.Manager
	cfg    Config
	client *http.Client
}

// New constructs a Gateway.
func New(jm *jwt.Manager, cfg Config) *Gateway {
	if cfg.CookieName == "" {
		cfg.CookieName = "access_token"
	}
	return &Gateway{
		jwt:    jm,
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// setCookie writes the auth cookie holding the access token.
func (g *Gateway) setCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     g.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   g.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(g.jwt.TTL().Seconds()),
	})
}

// clearCookie expires the auth cookie.
func (g *Gateway) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     g.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   g.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
