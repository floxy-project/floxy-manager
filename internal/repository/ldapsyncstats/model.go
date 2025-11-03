package ldapsyncstats

import (
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type ldapSyncStatsModel struct {
	ID            uint       `db:"id"`
	SyncSessionID string     `db:"sync_session_id"`
	StartTime     time.Time  `db:"start_time"`
	EndTime       *time.Time `db:"end_time"`
	Duration      *string    `db:"duration"`
	TotalUsers    int        `db:"total_users"`
	SyncedUsers   int        `db:"synced_users"`
	Errors        int        `db:"errors"`
	Warnings      int        `db:"warnings"`
	Status        string     `db:"status"`
	ErrorMessage  *string    `db:"error_message"`
}

func (m *ldapSyncStatsModel) toDomain() domain.LDAPSyncStats {
	return domain.LDAPSyncStats{
		ID:            m.ID,
		SyncSessionID: m.SyncSessionID,
		StartTime:     m.StartTime,
		EndTime:       m.EndTime,
		Duration:      m.Duration,
		TotalUsers:    m.TotalUsers,
		SyncedUsers:   m.SyncedUsers,
		Errors:        m.Errors,
		Warnings:      m.Warnings,
		Status:        m.Status,
		ErrorMessage:  m.ErrorMessage,
	}
}
