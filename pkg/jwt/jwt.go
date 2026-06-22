// Package jwt issues and verifies the HS256 access tokens used across services.
// The gateway verifies tokens and forwards identity to downstream services as
// trusted headers, so only the identity service and the gateway need the secret.
package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload carried by an access token.
type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwtlib.RegisteredClaims
}

// Manager signs and verifies tokens with a shared symmetric secret.
type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewManager builds a Manager. ttl is the access-token lifetime.
func NewManager(secret, issuer string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// TTL returns the configured token lifetime (used for cookie expiry).
func (m *Manager) TTL() time.Duration { return m.ttl }

// Issue mints a signed token for the given user.
func (m *Manager) Issue(userID, email, name, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)
}

// Verify parses and validates a token string, returning its claims.
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwtlib.ParseWithClaims(tokenStr, claims, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
