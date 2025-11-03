package domain

import (
	"time"
)

// LDAPSyncStatus represents the current status of LDAP synchronization.
type LDAPSyncStatus struct {
	Status           string    `json:"status"`
	IsRunning        bool      `json:"is_running"`
	LastSyncTime     time.Time `json:"last_sync_time"`
	TotalUsers       int       `json:"total_users"`
	SyncedUsers      int       `json:"synced_users"`
	Errors           int       `json:"errors"`
	Warnings         int       `json:"warnings"`
	LastSyncDuration string    `json:"last_sync_duration"`
}

// LDAPSyncProgress represents the progress of an ongoing LDAP synchronization.
type LDAPSyncProgress struct {
	IsRunning      bool      `json:"is_running"`
	Progress       float64   `json:"progress"` // 0-100
	CurrentStep    string    `json:"current_step"`
	ProcessedItems int       `json:"processed_items"`
	TotalItems     int       `json:"total_items"`
	EstimatedTime  string    `json:"estimated_time"`
	StartTime      time.Time `json:"start_time"`
	SyncID         string    `json:"sync_id"`
}

// LDAPSyncLog represents a log entry for LDAP synchronization.
type LDAPSyncLog struct {
	ID               uint      `db:"id"                 json:"id"`
	Timestamp        time.Time `db:"timestamp"          json:"timestamp"`
	Level            string    `db:"level"              json:"level"`
	Message          string    `db:"message"            json:"message"`
	Username         *string   `db:"username"           json:"username"`
	Details          *string   `db:"details"            json:"details"`
	SyncSessionID    string    `db:"sync_session_id"    json:"sync_session_id"`
	StackTrace       *string   `db:"stack_trace"        json:"stack_trace"`
	LDAPErrorCode    *int      `db:"ldap_error_code"    json:"ldap_error_code"`
	LDAPErrorMessage *string   `db:"ldap_error_message" json:"ldap_error_message"`
}

// LDAPSyncStats represents statistics for LDAP synchronization.
type LDAPSyncStats struct {
	ID            uint       `db:"id"              json:"id"`
	SyncSessionID string     `db:"sync_session_id" json:"sync_session_id"`
	StartTime     time.Time  `db:"start_time"      json:"start_time"`
	EndTime       *time.Time `db:"end_time"        json:"end_time"`
	Duration      *string    `db:"duration"        json:"duration"`
	TotalUsers    int        `db:"total_users"     json:"total_users"`
	SyncedUsers   int        `db:"synced_users"    json:"synced_users"`
	Errors        int        `db:"errors"          json:"errors"`
	Warnings      int        `db:"warnings"        json:"warnings"`
	Status        string     `db:"status"          json:"status"`
	ErrorMessage  *string    `db:"error_message"   json:"error_message"`
}

// LDAPSyncLogFilter represents filter parameters for LDAP sync logs.
type LDAPSyncLogFilter struct {
	Limit    *int       `json:"limit"`
	Level    *string    `json:"level"`
	SyncID   *string    `json:"sync_id"`
	Username *string    `json:"username"`
	From     *time.Time `json:"from"`
	To       *time.Time `json:"to"`
}

// LDAPSyncLogsResult represents paginated result of LDAP sync logs.
type LDAPSyncLogsResult struct {
	Logs    []LDAPSyncLog `json:"logs"`
	Total   int           `json:"total"`
	HasMore bool          `json:"has_more"`
}

// LDAPStatistics represents comprehensive LDAP statistics.
type LDAPStatistics struct {
	LDAPUsers       int                         `json:"ldap_users"`
	LocalUsers      int                         `json:"local_users"`
	ActiveUsers     int                         `json:"active_users"`
	InactiveUsers   int                         `json:"inactive_users"`
	SyncHistory     []LDAPStatisticsSyncHistory `json:"sync_history"`
	SyncSuccessRate float32                     `json:"sync_success_rate"`
}

// LDAPStatisticsSyncHistory represents sync history item.
type LDAPStatisticsSyncHistory struct {
	Date            time.Time `json:"date"`
	UsersSynced     int       `json:"users_synced"`
	Errors          int       `json:"errors"`
	DurationMinutes float32   `json:"duration_minutes"`
}
