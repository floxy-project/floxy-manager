package domain

import (
	"encoding/json"
	"time"
)

// Setting represents a configuration setting stored in the database.
type Setting struct {
	ID          int             `db:"id"          json:"id"`
	Name        string          `db:"name"        json:"name"`
	Value       json.RawMessage `db:"value"       json:"value"`
	Description string          `db:"description" json:"description"`
	CreatedAt   time.Time       `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"  json:"updated_at"`
}

// LDAPConfig represents LDAP configuration stored in settings.
type LDAPConfig struct {
	Enabled       bool   `json:"enabled"`
	URL           string `json:"url"`
	BindDN        string `json:"bind_dn"`
	BindPassword  string `json:"bind_password"`
	UserBaseDN    string `json:"user_base_dn"`
	UserFilter    string `json:"user_filter"`
	UserNameAttr  string `json:"user_name_attr"`
	UserEmailAttr string `json:"user_email_attr"`
	StartTLS      bool   `json:"start_tls"`
	InsecureTLS   bool   `json:"insecure_tls"`
	Timeout       string `json:"timeout"`
	SyncInterval  uint   `json:"sync_interval"`
}
