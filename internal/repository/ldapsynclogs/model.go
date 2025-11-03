package ldapsynclogs

import (
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type ldapSyncLogModel struct {
	ID               uint      `db:"id"`
	Timestamp        time.Time `db:"timestamp"`
	Level            string    `db:"level"`
	Message          string    `db:"message"`
	Username         *string   `db:"username"`
	Details          *string   `db:"details"`
	SyncSessionID    string    `db:"sync_session_id"`
	StackTrace       *string   `db:"stack_trace"`
	LDAPErrorCode    *int      `db:"ldap_error_code"`
	LDAPErrorMessage *string   `db:"ldap_error_message"`
}

func (m *ldapSyncLogModel) toDomain() domain.LDAPSyncLog {
	return domain.LDAPSyncLog{
		ID:               m.ID,
		Timestamp:        m.Timestamp,
		Level:            m.Level,
		Message:          m.Message,
		Username:         m.Username,
		Details:          m.Details,
		SyncSessionID:    m.SyncSessionID,
		StackTrace:       m.StackTrace,
		LDAPErrorCode:    m.LDAPErrorCode,
		LDAPErrorMessage: m.LDAPErrorMessage,
	}
}
