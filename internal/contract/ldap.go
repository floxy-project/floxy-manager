package contract

import (
	"context"

	"github.com/rom8726/floxy-manager/internal/domain"
)

// LDAPService defines the interface for LDAP service operations.
type LDAPService interface {
	// StartManualSync starts a manual LDAP synchronization of the specified type
	StartManualSync(ctx context.Context, syncID string, stopped chan struct{}) error
	// CancelSync cancels an ongoing synchronization
	CancelSync(ctx context.Context) error
	// GetSyncStatus returns the current status of LDAP synchronization
	GetSyncStatus() domain.LDAPSyncStatus
	// GetSyncProgress returns the progress of an ongoing LDAP synchronization
	GetSyncProgress() domain.LDAPSyncProgress
	// TestConnection tests the connection to the LDAP server
	TestConnection(ctx context.Context) error
	// ReloadConfig reloads config and creates new LDAP client
	ReloadConfig(ctx context.Context) error
	// SyncUsers synchronizes users from LDAP
	SyncUsers(ctx context.Context) error
}

type LDAPSyncLogsRepository interface {
	Create(ctx context.Context, log domain.LDAPSyncLog) (domain.LDAPSyncLog, error)
	GetByID(ctx context.Context, id uint) (domain.LDAPSyncLog, error)
	List(ctx context.Context, filter domain.LDAPSyncLogFilter) (domain.LDAPSyncLogsResult, error)
	DeleteBySyncID(ctx context.Context, syncSessionID string) error
}

type LDAPSyncStatsRepository interface {
	Create(ctx context.Context, stats domain.LDAPSyncStats) (domain.LDAPSyncStats, error)
	GetBySyncSessionID(ctx context.Context, syncSessionID string) (domain.LDAPSyncStats, error)
	Update(ctx context.Context, stats domain.LDAPSyncStats) error
	List(ctx context.Context, limit int) ([]domain.LDAPSyncStats, error)
	GetStatistics(ctx context.Context) (domain.LDAPStatistics, error)
}

type LDAPSyncUseCase interface {
	GetSyncStatus(ctx context.Context) (domain.LDAPSyncStatus, error)
	GetSyncProgress(ctx context.Context) (domain.LDAPSyncProgress, error)
	StartManualSync(ctx context.Context) error
	CancelSync(ctx context.Context) error
	TestConnection(ctx context.Context) error
	GetSyncLogs(ctx context.Context, filter domain.LDAPSyncLogFilter) (domain.LDAPSyncLogsResult, error)
	GetSyncLogDetails(ctx context.Context, id uint) (domain.LDAPSyncLog, error)
	GetStatistics(ctx context.Context) (domain.LDAPStatistics, error)
	UpdateConfig(ctx context.Context, cfg *domain.LDAPConfig) error
	ReloadConfig(ctx context.Context) error
	Disable(ctx context.Context) error
}
