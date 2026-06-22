package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"swiftmind/pkg/domain"
	"swiftmind/pkg/httpx"
)

// Router builds the full gateway HTTP handler.
func (g *Gateway) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(g.cors)

	r.Get("/health", httpx.Health)

	r.Route("/api", func(r chi.Router) {
		// Public auth endpoints.
		r.Post("/auth/login", g.login)
		r.Post("/auth/logout", g.logout)

		// Authenticated endpoints.
		r.Group(func(r chi.Router) {
			r.Use(g.authRequired)
			r.Get("/auth/me", g.me)

			// Fine rules: any authenticated user may read the active ruleset and
			// version history; only officers may publish a new version.
			rulesProxy := g.proxyTo(g.cfg.RulesURL, "/api/rules")
			r.Route("/rules", func(r chi.Router) {
				r.Get("/active", rulesProxy.ServeHTTP)
				r.Get("/", rulesProxy.ServeHTTP)
				r.With(g.requireRole(domain.RoleOfficer)).Post("/", rulesProxy.ServeHTTP)
			})

			// Violations: officers submit; officers see all, members see their
			// own; the photo streams back through the service.
			violationProxy := g.proxyTo(g.cfg.ViolationURL, "/api/violations")
			r.Route("/violations", func(r chi.Router) {
				r.Get("/", violationProxy.ServeHTTP)
				r.Get("/{id}/photo", violationProxy.ServeHTTP)
				r.With(g.requireRole(domain.RoleOfficer)).Post("/", violationProxy.ServeHTTP)
			})

			// Invoices: officers and members may list (scoped); only members pay.
			paymentProxy := g.proxyTo(g.cfg.PaymentURL, "/api/invoices")
			r.Route("/invoices", func(r chi.Router) {
				r.Get("/", paymentProxy.ServeHTTP)
				r.With(g.requireRole(domain.RoleMember)).Post("/{id}/pay", paymentProxy.ServeHTTP)
			})
		})
	})

	return r
}

// cors permits the browser frontend (when it calls the gateway directly during
// development) to send the auth cookie.
func (g *Gateway) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := g.cfg.AllowOrigin
		if origin == "" {
			origin = "http://localhost:3000"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// login forwards credentials to the identity service and, on success, sets the
// httpOnly auth cookie so the browser is authenticated for subsequent calls.
func (g *Gateway) login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, g.cfg.IdentityURL+"/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "identity service unavailable")
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		return
	}

	var parsed struct {
		Token string          `json:"token"`
		User  json.RawMessage `json:"user"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil || parsed.Token == "" {
		httpx.Error(w, http.StatusBadGateway, "unexpected identity response")
		return
	}

	g.setCookie(w, parsed.Token)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": parsed.User})
}

// logout clears the auth cookie.
func (g *Gateway) logout(w http.ResponseWriter, _ *http.Request) {
	g.clearCookie(w)
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// me returns the current user, derived from the verified token claims.
func (g *Gateway) me(w http.ResponseWriter, r *http.Request) {
	claims := claimsFromCtx(r.Context())
	if claims == nil {
		httpx.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"id":    claims.UserID,
			"email": claims.Email,
			"name":  claims.Name,
			"role":  claims.Role,
		},
	})
}
