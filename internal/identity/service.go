package identity

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"swiftmind/pkg/domain"
	"swiftmind/pkg/jwt"
)

// ErrInvalidCredentials is returned for unknown email or wrong password.
var ErrInvalidCredentials = errors.New("invalid email or password")

// Service holds the identity business logic.
type Service struct {
	store *Store
	jwt   *jwt.Manager
}

// NewService wires the store and the JWT manager.
func NewService(store *Store, jm *jwt.Manager) *Service {
	return &Service{store: store, jwt: jm}
}

// PublicUser is the user shape safe to return over the API.
type PublicUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// LoginResult bundles the signed token and the authenticated user.
type LoginResult struct {
	Token string
	User  PublicUser
}

type seedUser struct {
	email, password, name, role string
}

// Seed inserts the demo officer and member on first start (empty users table).
// Credentials are documented in the README.
func (s *Service) Seed(ctx context.Context) error {
	n, err := s.store.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	seeds := []seedUser{
		{"officer@swiftmind.test", "password123", "Olivia Officer", domain.RoleOfficer.String()},
		{"member@swiftmind.test", "password123", "Mike Member", domain.RoleMember.String()},
	}
	for _, su := range seeds {
		hash, err := bcrypt.GenerateFromPassword([]byte(su.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := s.store.Insert(ctx, su.email, string(hash), su.name, su.role); err != nil {
			return err
		}
	}
	return nil
}

// Login validates credentials and issues an access token.
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	u, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	token, err := s.jwt.Issue(u.ID, u.Email, u.Name, u.Role)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token: token,
		User:  PublicUser{ID: u.ID, Email: u.Email, Name: u.Name, Role: u.Role},
	}, nil
}
