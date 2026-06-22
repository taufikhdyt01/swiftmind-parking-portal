package gateway

import (
	"context"
	"net/http"
	"strings"

	"swiftmind/pkg/domain"
	"swiftmind/pkg/httpx"
	"swiftmind/pkg/jwt"
)

type ctxKey string

const claimsKey ctxKey = "claims"

// tokenFromRequest extracts the JWT from the auth cookie or a Bearer header.
func (g *Gateway) tokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(g.cfg.CookieName); err == nil && c.Value != "" {
		return c.Value
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// authRequired rejects unauthenticated requests and stashes the claims in context.
func (g *Gateway) authRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := g.tokenFromRequest(r)
		if tok == "" {
			httpx.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		claims, err := g.jwt.Verify(tok)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireRole guards a route so only the given role may pass. Used by later phases
// (e.g. only officers may publish rules or submit violations).
func (g *Gateway) requireRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := claimsFromCtx(r.Context())
			if claims == nil || claims.Role != role.String() {
				httpx.Error(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// claimsFromCtx returns the verified claims placed by authRequired, or nil.
func claimsFromCtx(ctx context.Context) *jwt.Claims {
	c, _ := ctx.Value(claimsKey).(*jwt.Claims)
	return c
}
