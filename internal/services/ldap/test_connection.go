package ldap

import (
	"context"
	"fmt"

	ldap "github.com/go-ldap/ldap/v3"
)

// TestConnection It attempts to bind with the configured credentials and perform a simple search.
func (c *Client) TestConnection(ctx context.Context) error {
	return c.withConnection(func(conn *ldap.Conn) error {
		// If we got here, the connection and bind were successful
		// to Perform a simple search to verify search permissions
		searchRequest := ldap.NewSearchRequest(
			c.config.UserBaseDN,
			ldap.ScopeBaseObject,
			ldap.NeverDerefAliases, 0, 1, false,
			"(objectClass=*)",
			[]string{"1.1"}, // Request no attributes
			nil,
		)

		_, err := conn.Search(searchRequest)
		if err != nil {
			return fmt.Errorf("LDAP search test failed: %w", err)
		}

		return nil
	})
}
