// Package domain holds cross-service primitives shared by every service:
// roles and the trusted identity headers the gateway forwards downstream.
package domain

// Role identifies a portal user's role.
type Role string

const (
	RoleOfficer Role = "officer"
	RoleMember  Role = "member"
)

// Valid reports whether r is a recognized role.
func (r Role) Valid() bool {
	return r == RoleOfficer || r == RoleMember
}

// String returns the role as a plain string.
func (r Role) String() string { return string(r) }

// Trusted identity headers. The gateway sets these after verifying the JWT;
// downstream services read them and must only be reachable via the gateway.
const (
	HeaderUserID    = "X-User-Id"
	HeaderUserRole  = "X-User-Role"
	HeaderUserEmail = "X-User-Email"
	HeaderUserName  = "X-User-Name"
)
