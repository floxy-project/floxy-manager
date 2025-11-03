package ldap

import (
	"context"
)

// ClientService defines the interface for LDAP client operations.
type ClientService interface {
	// Authenticate authenticates a user against the LDAP server
	Authenticate(ctx context.Context, username, password string) (bool, error)

	// GetUser retrieves user attributes from LDAP
	GetUser(ctx context.Context, username string) (map[string][]string, error)

	// GetAllUsers retrieves all usernames from LDAP
	GetAllUsers(ctx context.Context) ([]string, error)

	// TestConnection tests the connection to the LDAP server
	TestConnection(ctx context.Context) error

	// Close closes the LDAP connection and releases resources
	Close() error
}
