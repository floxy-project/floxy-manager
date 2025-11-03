//nolint:gosec,gocyclo,nestif // need refactoring
package ldap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

// Client represents an LDAP client that implements the ClientService interface.
type Client struct {
	config   *ClientConfig
	connPool chan *ldap.Conn
	mu       sync.RWMutex
	closed   bool

	// Connection health tracking
	connCreatedAt map[*ldap.Conn]time.Time
	healthTicker  *time.Ticker
	healthStop    chan struct{}
}

// ClientConfig holds the configuration for the LDAP client.
type ClientConfig struct {
	// Connection settings
	URL          string        `mapstructure:"url"`
	StartTLS     bool          `mapstructure:"start_tls"`
	InsecureTLS  bool          `mapstructure:"insecure_tls"`
	Timeout      time.Duration `mapstructure:"timeout"`
	BindDN       string        `mapstructure:"bind_dn"`
	BindPassword string        `mapstructure:"bind_password"`

	// User settings
	UserBaseDN    string `mapstructure:"user_base_dn"`
	UserFilter    string `mapstructure:"user_filter"`
	UserNameAttr  string `mapstructure:"user_name_attr"`
	UserEmailAttr string `mapstructure:"user_email_attr"`

	// Connection pooling
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// NewClient creates a new LDAP client with connection pooling.
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		return nil, errors.New("LDAP client config cannot be nil")
	}

	// Set default values
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 10
	}

	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}

	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 5 * time.Minute
	}

	client := &Client{
		config:        config,
		connPool:      make(chan *ldap.Conn, config.MaxOpenConns),
		connCreatedAt: make(map[*ldap.Conn]time.Time),
		healthStop:    make(chan struct{}),
	}

	// Initialize connection pool
	for range config.MaxIdleConns {
		conn, err := client.createConnection()
		if err != nil {
			_ = client.Close()

			return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
		}
		client.connPool <- conn
		client.connCreatedAt[conn] = time.Now()
	}

	// Start connection health monitoring
	client.startHealthMonitoring()

	return client, nil
}

// Authenticate authenticates a user against the LDAP server.
func (c *Client) Authenticate(_ context.Context, username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, errors.New("username and password are required")
	}

	var authenticated bool

	err := c.withConnection(func(conn *ldap.Conn) error {
		// Search for the user
		searchRequest := ldap.NewSearchRequest(
			c.config.UserBaseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", c.config.UserFilter, c.config.UserNameAttr, ldap.EscapeFilter(username)),
			[]string{"dn"},
			nil,
		)

		searchResult, err := conn.Search(searchRequest)
		if err != nil {
			return fmt.Errorf("LDAP search failed: %w", err)
		}

		if len(searchResult.Entries) == 0 {
			return nil // User not found
		}

		if len(searchResult.Entries) > 1 {
			return fmt.Errorf("multiple users found with username: %s", username)
		}

		userDN := searchResult.Entries[0].DN

		// Try to bind as the user to verify the password
		err = conn.Bind(userDN, password)
		if err != nil {
			if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
				return nil // Invalid credentials, not an error
			}

			return fmt.Errorf("LDAP bind failed: %w", err)
		}

		authenticated = true

		return nil
	})

	return authenticated, err
}

// GetUser retrieves user attributes from LDAP.
func (c *Client) GetUser(_ context.Context, username string) (map[string][]string, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	var attributes map[string][]string

	err := c.withConnection(func(conn *ldap.Conn) error {
		searchRequest := ldap.NewSearchRequest(
			c.config.UserBaseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", c.config.UserFilter, c.config.UserNameAttr, ldap.EscapeFilter(username)),
			[]string{"*"},
			nil,
		)

		searchResult, err := conn.Search(searchRequest)
		if err != nil {
			return fmt.Errorf("LDAP search failed: %w", err)
		}

		if len(searchResult.Entries) == 0 {
			return fmt.Errorf("user not found: %s", username)
		}

		if len(searchResult.Entries) > 1 {
			return fmt.Errorf("multiple users found with username: %s", username)
		}

		entry := searchResult.Entries[0]
		attributes = make(map[string][]string)

		for _, attr := range entry.Attributes {
			attributes[attr.Name] = attr.Values
		}

		return nil
	})

	return attributes, err
}

// GetAllUsers retrieves all usernames from LDAP.
func (c *Client) GetAllUsers(_ context.Context) ([]string, error) {
	var usernames []string

	err := c.withConnection(func(conn *ldap.Conn) error {
		searchRequest := ldap.NewSearchRequest(
			c.config.UserBaseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases, 0, 0, false,
			c.config.UserFilter,
			[]string{c.config.UserNameAttr},
			nil,
		)

		searchResult, err := conn.Search(searchRequest)
		if err != nil {
			return fmt.Errorf("LDAP search failed: %w", err)
		}

		usernames = make([]string, 0, len(searchResult.Entries))

		for _, entry := range searchResult.Entries {
			for _, attr := range entry.Attributes {
				if attr.Name == c.config.UserNameAttr && len(attr.Values) > 0 {
					usernames = append(usernames, attr.Values[0])

					break
				}
			}
		}

		return nil
	})

	return usernames, err
}

// Close closes all connections in the pool and cleans up resources.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Stop health monitoring
	if c.healthTicker != nil {
		c.healthTicker.Stop()
	}

	close(c.healthStop)

	// Close the connection pool channel
	close(c.connPool)

	var errs []error

	for conn := range c.connPool {
		if conn != nil {
			if err := conn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Clear the connection tracking map
	c.connCreatedAt = nil

	if len(errs) > 0 {
		return fmt.Errorf("errors while closing connections: %v", errs)
	}

	return nil
}

// createConnection creates a new LDAP connection.
func (c *Client) createConnection() (*ldap.Conn, error) {
	// Create new connection
	var conn *ldap.Conn

	var err error

	if strings.HasPrefix(c.config.URL, "ldaps://") {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.config.InsecureTLS,
		}
		conn, err = ldap.DialURL(strings.TrimPrefix(c.config.URL, "ldaps://"), ldap.DialWithTLSConfig(tlsConfig))
	} else {
		conn, err = ldap.DialURL(
			c.config.URL,
			ldap.DialWithDialer(&net.Dialer{Timeout: c.config.Timeout}),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	// Set timeouts
	conn.SetTimeout(c.config.Timeout)

	// Configure StartTLS if needed
	if c.config.StartTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.config.InsecureTLS,
		}
		if err := conn.StartTLS(tlsConfig); err != nil {
			_ = conn.Close()

			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Bind with the provided credentials
	if c.config.BindDN != "" && c.config.BindPassword != "" {
		if err := conn.Bind(c.config.BindDN, c.config.BindPassword); err != nil {
			_ = conn.Close()

			return nil, fmt.Errorf("failed to bind to LDAP server: %w", err)
		}
	}

	// Track connection creation time
	c.connCreatedAt[conn] = time.Now()

	return conn, nil
}

// getConnection gets a connection from the pool or creates a new one.
func (c *Client) getConnection() (*ldap.Conn, error) {
	if c.closed {
		return nil, errors.New("LDAP client is closed")
	}

	select {
	case conn := <-c.connPool:
		// Check if the connection is still valid
		if conn.IsClosing() {
			_ = conn.Close()
			delete(c.connCreatedAt, conn)

			return c.createConnection()
		}

		// Additional check: try to perform a simple operation to verify connection health
		if err := c.testConnectionHealth(conn); err != nil {
			_ = conn.Close()
			delete(c.connCreatedAt, conn)

			return c.createConnection()
		}

		return conn, nil
	default:
		// No available connection, create a new one if under max
		c.mu.Lock()
		defer c.mu.Unlock()

		// Check again after acquiring the lock
		select {
		case conn := <-c.connPool:
			// Check if the connection is still valid
			if conn.IsClosing() {
				_ = conn.Close()
				delete(c.connCreatedAt, conn)

				return c.createConnection()
			}

			// Additional check: try to perform a simple operation to verify connection health
			if err := c.testConnectionHealth(conn); err != nil {
				_ = conn.Close()
				delete(c.connCreatedAt, conn)

				return c.createConnection()
			}

			return conn, nil
		default:
			// Check if we can create a new connection
			if len(c.connPool) >= c.config.MaxOpenConns {
				return nil, errors.New("max connections reached")
			}

			return c.createConnection()
		}
	}
}

// testConnectionHealth performs a simple LDAP operation to verify the connection is still healthy.
func (c *Client) testConnectionHealth(conn *ldap.Conn) error {
	// Perform a simple search to test connection health
	searchRequest := ldap.NewSearchRequest(
		c.config.UserBaseDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases, 0, 1, false,
		"(objectClass=*)",
		[]string{"1.1"}, // Request no attributes
		nil,
	)

	_, err := conn.Search(searchRequest)

	return err
}

// releaseConnection returns a connection to the pool.
func (c *Client) releaseConnection(conn *ldap.Conn) {
	if c.closed || conn == nil || conn.IsClosing() {
		if conn != nil {
			_ = conn.Close()
			delete(c.connCreatedAt, conn)
		}

		return
	}

	select {
	case c.connPool <- conn:
		// Connection returned to pool
	default:
		// Pool is full, close the connection
		_ = conn.Close()
		delete(c.connCreatedAt, conn)
	}
}

// withConnection executes a function with a connection from the pool.
func (c *Client) withConnection(fn func(conn *ldap.Conn) error) error {
	// Get a connection from the pool
	conn, err := c.getConnection()
	if err != nil {
		err = fmt.Errorf("failed to get LDAP connection: %w", err)

		return err
	}
	defer c.releaseConnection(conn)

	// Execute the function with the connection
	err = fn(conn)
	if err != nil {
		// Check if this is a connection-related error that requires reconnection
		if isConnectionError(err) {
			slog.Warn("LDAP connection error detected, attempting reconnection", "error", err)

			// Close the problematic connection and try again with a fresh one
			_ = conn.Close()
			delete(c.connCreatedAt, conn)

			// Get a fresh connection
			freshConn, freshErr := c.getConnection()
			if freshErr != nil {
				return fmt.Errorf("failed to get fresh LDAP connection after error: %w", freshErr)
			}
			defer c.releaseConnection(freshConn)

			// Try the operation again with the fresh connection
			err = fn(freshConn)
			if err != nil {
				slog.Error("LDAP operation failed even with fresh connection", "error", err)
			} else {
				slog.Info("LDAP operation succeeded with fresh connection after reconnection")
			}
		} else {
			slog.Error("LDAP operation failed", "error", err)
		}
	}

	return err
}

// isConnectionError checks if the error indicates a connection problem that requires reconnection.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	var ldapErr *ldap.Error
	// Check for specific LDAP error codes that indicate connection issues
	if errors.As(err, &ldapErr) {
		switch ldapErr.ResultCode {
		case ldap.LDAPResultNoSuchObject, // 32
			ldap.LDAPResultUnavailable,           // 52
			ldap.LDAPResultServerDown,            // 81
			ldap.LDAPResultLocalError,            // 82
			ldap.LDAPResultEncodingError,         // 83
			ldap.LDAPResultDecodingError,         // 84
			ldap.LDAPResultTimeout,               // 85
			ldap.LDAPResultAuthUnknown,           // 86
			ldap.LDAPResultFilterError,           // 87
			ldap.LDAPResultUserCanceled,          // 88
			ldap.LDAPResultParamError,            // 89
			ldap.LDAPResultNoMemory,              // 90
			ldap.LDAPResultConnectError,          // 91
			ldap.LDAPResultNotSupported,          // 92
			ldap.LDAPResultControlNotFound,       // 93
			ldap.LDAPResultNoResultsReturned,     // 94
			ldap.LDAPResultMoreResultsToReturn,   // 95
			ldap.LDAPResultClientLoop,            // 96
			ldap.LDAPResultReferralLimitExceeded, // 97
			ldap.LDAPResultInvalidResponse,       // 100
			ldap.LDAPResultAmbiguousResponse,     // 101
			ldap.LDAPResultTLSNotSupported,       // 112
			ldap.LDAPResultIntermediateResponse,  // 113
			// ldap.LDAPResultUnknownResponse,   // 114
			// ldap.LDAPResultUnknownSaslCredentials, // 115
			ldap.LDAPResultSaslBindInProgress: // 116
			return true
		}
	}

	// Check for network-related errors
	errStr := err.Error()
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "use of closed network connection") {
		return true
	}

	return false
}

// startHealthMonitoring starts a background goroutine to monitor connection health.
func (c *Client) startHealthMonitoring() {
	// Check connections every 30 seconds
	c.healthTicker = time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-c.healthTicker.C:
				c.cleanupStaleConnections()
			case <-c.healthStop:
				return
			}
		}
	}()
}

// cleanupStaleConnections removes stale connections from the pool.
func (c *Client) cleanupStaleConnections() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if a client is closed
	if c.closed {
		return
	}

	// Get the current pool size
	currentPoolSize := len(c.connPool)
	if currentPoolSize == 0 {
		return
	}

	slog.Debug("Starting LDAP connection pool cleanup", "pool_size", currentPoolSize)

	// Create a temporary slice to hold valid connections
	validConns := make([]*ldap.Conn, 0, currentPoolSize)
	staleCount := 0

	// Drain the pool and check each connection
poolLoop:
	for range currentPoolSize {
		select {
		case conn := <-c.connPool:
			// Check if the connection is too old or unhealthy
			if c.isConnectionStale(conn) {
				_ = conn.Close()
				delete(c.connCreatedAt, conn)
				staleCount++
				slog.Debug("Removed stale LDAP connection from pool")
			} else {
				validConns = append(validConns, conn)
			}
		default:
			slog.Debug("LDAP connection pool is empty, skipping cleanup")

			break poolLoop
		}
	}

	// Return valid connections to the pool
	for _, conn := range validConns {
		select {
		case c.connPool <- conn:
		default:
			// Pool is full, close the connection
			_ = conn.Close()
			delete(c.connCreatedAt, conn)
		}
	}

	// Replenish the pool if needed
	targetSize := c.config.MaxIdleConns
	currentSize := len(c.connPool)

	for i := currentSize; i < targetSize; i++ {
		conn, err := c.createConnection()
		if err != nil {
			slog.Error("Failed to create replacement connection", "error", err)

			break
		}

		select {
		case c.connPool <- conn:
			c.connCreatedAt[conn] = time.Now()
		default:
			_ = conn.Close()
		}
	}

	slog.Debug("Completed LDAP connection pool cleanup",
		"stale_removed", staleCount,
		"final_pool_size", len(c.connPool))
}

// isConnectionStale checks if a connection is too old or unhealthy.
func (c *Client) isConnectionStale(conn *ldap.Conn) bool {
	// Check if the connection is closing
	if conn.IsClosing() {
		return true
	}

	// Check if the connection is too old
	if createdAt, exists := c.connCreatedAt[conn]; exists {
		if time.Since(createdAt) > c.config.ConnMaxLifetime {
			return true
		}
	}

	// Test connection health
	if err := c.testConnectionHealth(conn); err != nil {
		return true
	}

	return false
}
