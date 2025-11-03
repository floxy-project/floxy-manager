# LDAP Service

This package provides LDAP authentication and user synchronization functionality for the Floxy Manager application.

## Features

- LDAP authentication with configurable connection settings
- User attribute synchronization from LDAP to local database
- Connection pooling for better performance
- TLS/StartTLS support
- OpenTelemetry integration for observability

## Configuration

The LDAP client can be configured using the following environment variables or configuration file:

```yaml
ldap:
  # Connection settings
  url: "ldap://ldap.example.com:389"  # LDAP server URL (ldap:// or ldaps://)
  start_tls: false                   # Use StartTLS
  insecure_tls: false                 # Skip TLS certificate verification
  timeout: 30s                        # Connection timeout
  bind_dn: "cn=admin,dc=example,dc=com"
  bind_password: "secret"
  
  # User settings
  user_base_dn: "ou=users,dc=example,dc=com"
  user_filter: "(objectClass=person)"
  user_name_attr: "uid"
  user_email_attr: "mail"
  
  # Connection pooling
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 5m
```

## Usage

### Creating a Client

```go
import (
    "github.com/rom8726/floxy-manager/internal/services/ldap"
    "go.uber.org/zap"
)

config := &ldap.ClientConfig{
    URL:          "ldap://ldap.example.com:389",
    BindDN:       "cn=admin,dc=example,dc=com",
    BindPassword: "secret",
    UserBaseDN:   "ou=users,dc=example,dc=com",
    UserFilter:   "(objectClass=person)",
    UserNameAttr: "uid",
}

logger := zap.NewNop() // or your configured logger
client, err := ldap.NewClient(config, logger)
if err != nil {
    // handle error
}
defer client.Close()
```

### Authenticating a User

```go
authenticated, err := client.Authenticate(context.Background(), "username", "password")
if err != nil {
    // handle error
}

if authenticated {
    // User authenticated successfully
}
```

### Getting User Details

```go
attrs, err := client.GetUser(context.Background(), "username")
if err != nil {
    // handle error
}

// Access user attributes
email := attrs["mail"]
```

## Testing

Run the tests with:

```bash
go test -v ./internal/services/ldap/...
```

## Dependencies

- `github.com/go-ldap/ldap/v3` - LDAP client library
- `go.opentelemetry.io/otel` - OpenTelemetry for distributed tracing
- `go.uber.org/zap` - Structured logging

## Security Considerations

- Always use TLS/SSL (ldaps:// or StartTLS) in production
- Store LDAP bind credentials securely (e.g., using a secret manager)
- Implement proper access controls on the LDAP server
- Set appropriate timeouts to prevent hanging connections
- Rotate credentials regularly
