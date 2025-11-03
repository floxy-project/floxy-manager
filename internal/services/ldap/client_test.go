package ldap

import (
	"errors"
	"testing"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "LDAP NoSuchObject error",
			err:      &ldap.Error{ResultCode: ldap.LDAPResultNoSuchObject},
			expected: true,
		},
		{
			name:     "LDAP ServerDown error",
			err:      &ldap.Error{ResultCode: ldap.LDAPResultServerDown},
			expected: true,
		},
		{
			name:     "LDAP Timeout error",
			err:      &ldap.Error{ResultCode: ldap.LDAPResultTimeout},
			expected: true,
		},
		{
			name:     "network connection error",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "broken pipe error",
			err:      errors.New("broken pipe"),
			expected: true,
		},
		{
			name:     "use of closed network connection",
			err:      errors.New("use of closed network connection"),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isConnectionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClientConnectionHealth(t *testing.T) {
	t.Skip("Skipping test that requires LDAP server")

	config := &ClientConfig{
		URL:          "ldap://localhost:389",
		BindDN:       "cn=admin,dc=example,dc=com",
		BindPassword: "password",
		UserBaseDN:   "ou=users,dc=example,dc=com",
		UserFilter:   "(objectClass=person)",
		UserNameAttr: "uid",
		Timeout:      5 * time.Second,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Skipf("Skipping test - cannot create LDAP client: %v", err)
	}
	defer client.Close()

	// Test that we can create connections
	conn, err := client.getConnection()
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Test connection health check
	err = client.testConnectionHealth(conn)
	// This might fail if LDAP server is not available, which is expected
	if err != nil {
		t.Logf("Connection health check failed (expected if no LDAP server): %v", err)
	}

	client.releaseConnection(conn)
}

func TestClientConfigDefaults(t *testing.T) {
	t.Skip("Skipping test that requires LDAP server")

	config := &ClientConfig{
		URL: "ldap://localhost:389",
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	// Check that defaults are set
	assert.Equal(t, 30*time.Second, client.config.Timeout)
	assert.Equal(t, 10, client.config.MaxOpenConns)
	assert.Equal(t, 5, client.config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, client.config.ConnMaxLifetime)
}

func TestClientClose(t *testing.T) {
	t.Skip("Skipping test that requires LDAP server")

	config := &ClientConfig{
		URL: "ldap://localhost:389",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	// Test that close works without error
	err = client.Close()
	assert.NoError(t, err)

	// Test that close is idempotent
	err = client.Close()
	assert.NoError(t, err)
}
