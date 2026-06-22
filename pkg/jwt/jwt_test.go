package jwt

import (
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	m := NewManager("secret", "swiftmind", time.Hour)

	token, err := m.Issue("u1", "alice@swiftmind.test", "Alice", "officer")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.UserID != "u1" {
		t.Errorf("UserID = %q, want u1", claims.UserID)
	}
	if claims.Email != "alice@swiftmind.test" {
		t.Errorf("Email = %q, want alice@swiftmind.test", claims.Email)
	}
	if claims.Role != "officer" {
		t.Errorf("Role = %q, want officer", claims.Role)
	}
	if claims.Issuer != "swiftmind" {
		t.Errorf("Issuer = %q, want swiftmind", claims.Issuer)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	signer := NewManager("secret-a", "swiftmind", time.Hour)
	verifier := NewManager("secret-b", "swiftmind", time.Hour)

	token, _ := signer.Issue("u1", "a@b.c", "Alice", "member")
	if _, err := verifier.Verify(token); err == nil {
		t.Fatal("expected verification to fail with a different secret")
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	m := NewManager("secret", "swiftmind", -time.Minute) // already expired

	token, _ := m.Issue("u1", "a@b.c", "Alice", "officer")
	if _, err := m.Verify(token); err == nil {
		t.Fatal("expected an expired token to be rejected")
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	m := NewManager("secret", "swiftmind", time.Hour)
	if _, err := m.Verify("not-a-jwt"); err == nil {
		t.Fatal("expected malformed token to be rejected")
	}
}
