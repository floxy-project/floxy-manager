package ldap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type UseCase struct {
	ldapService       contract.LDAPService
	ldapSyncLogsRepo  contract.LDAPSyncLogsRepository
	ldapSyncStatsRepo contract.LDAPSyncStatsRepository
	settingsUseCase   contract.SettingsUseCase
}

func New(
	ldapService contract.LDAPService,
	ldapSyncLogsRepo contract.LDAPSyncLogsRepository,
	ldapSyncStatsRepo contract.LDAPSyncStatsRepository,
	settingsUseCase contract.SettingsUseCase,
) *UseCase {
	return &UseCase{
		ldapService:       ldapService,
		ldapSyncLogsRepo:  ldapSyncLogsRepo,
		ldapSyncStatsRepo: ldapSyncStatsRepo,
		settingsUseCase:   settingsUseCase,
	}
}

// GetSyncStatus returns the current status of LDAP synchronization.
func (uc *UseCase) GetSyncStatus(ctx context.Context) (domain.LDAPSyncStatus, error) {
	// Get the current sync status from memory (for is_running)
	memoryStatus := uc.ldapService.GetSyncStatus()

	// If sync is currently running, return memory status
	if memoryStatus.IsRunning {
		return memoryStatus, nil
	}

	// Otherwise, get the last completed sync from a database
	stats, err := uc.ldapSyncStatsRepo.List(ctx, 30)
	if err != nil {
		// If no stats found, return default status
		return domain.LDAPSyncStatus{}, nil //nolint:nilerr // it's ok
	}

	if len(stats) == 0 {
		// No sync history found
		return domain.LDAPSyncStatus{}, nil
	}

	lastSync := stats[0]

	return domain.LDAPSyncStatus{
		Status:       lastSync.Status,
		IsRunning:    false,
		LastSyncTime: lastSync.StartTime,
		TotalUsers:   lastSync.TotalUsers,
		SyncedUsers:  lastSync.SyncedUsers,
		Errors:       lastSync.Errors,
		Warnings:     lastSync.Warnings,
		LastSyncDuration: func() string {
			if lastSync.Duration != nil {
				return *lastSync.Duration
			}

			return ""
		}(),
	}, nil
}

// GetSyncProgress returns the progress of an ongoing LDAP synchronization.
func (uc *UseCase) GetSyncProgress(context.Context) (domain.LDAPSyncProgress, error) {
	return uc.ldapService.GetSyncProgress(), nil
}

// StartManualSync starts a manual LDAP synchronization.
func (uc *UseCase) StartManualSync(ctx context.Context) error {
	// Create initial sync stats
	progress := uc.ldapService.GetSyncProgress()
	if progress.IsRunning {
		return errors.New("a sync is already running")
	}

	stats := domain.LDAPSyncStats{
		SyncSessionID: uuid.NewString(),
		StartTime:     time.Now(),
		Status:        "running",
	}

	_, err := uc.ldapSyncStatsRepo.Create(ctx, stats)
	if err != nil {
		return fmt.Errorf("failed to create sync stats: %w", err)
	}

	// Log sync start
	log := domain.LDAPSyncLog{
		Timestamp:     time.Now(),
		Level:         "info",
		Message:       "Starting synchronization",
		SyncSessionID: stats.SyncSessionID,
	}

	_, err = uc.ldapSyncLogsRepo.Create(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to create sync log: %w", err)
	}

	// Start the actual sync
	return uc.ldapService.StartManualSync(ctx, stats.SyncSessionID, nil)
}

// CancelSync cancels an ongoing synchronization.
func (uc *UseCase) CancelSync(ctx context.Context) error {
	progress := uc.ldapService.GetSyncProgress()
	if !progress.IsRunning {
		return errors.New("no sync is currently running")
	}

	// Update stats
	stats, err := uc.ldapSyncStatsRepo.GetBySyncSessionID(ctx, progress.SyncID)
	if err == nil {
		stats.Status = "cancelled"
		stats.EndTime = &time.Time{}
		duration := time.Since(progress.StartTime).String()
		stats.Duration = &duration

		err = uc.ldapSyncStatsRepo.Update(ctx, stats)
		if err != nil {
			return fmt.Errorf("failed to update sync stats: %w", err)
		}
	}

	// Log cancellation
	log := domain.LDAPSyncLog{
		Timestamp:     time.Now(),
		Level:         "warning",
		Message:       "Synchronization cancelled by user",
		SyncSessionID: progress.SyncID,
	}

	_, err = uc.ldapSyncLogsRepo.Create(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to create sync log: %w", err)
	}

	return uc.ldapService.CancelSync(ctx)
}

// TestConnection tests the connection to the LDAP server.
func (uc *UseCase) TestConnection(ctx context.Context) error {
	return uc.ldapService.TestConnection(ctx)
}

// GetSyncLogs returns LDAP synchronization logs with filtering.
func (uc *UseCase) GetSyncLogs(
	ctx context.Context,
	filter domain.LDAPSyncLogFilter,
) (domain.LDAPSyncLogsResult, error) {
	return uc.ldapSyncLogsRepo.List(ctx, filter)
}

// GetSyncLogDetails returns details of a specific sync log entry.
func (uc *UseCase) GetSyncLogDetails(ctx context.Context, id uint) (domain.LDAPSyncLog, error) {
	return uc.ldapSyncLogsRepo.GetByID(ctx, id)
}

// GetStatistics returns comprehensive LDAP statistics.
func (uc *UseCase) GetStatistics(ctx context.Context) (domain.LDAPStatistics, error) {
	return uc.ldapSyncStatsRepo.GetStatistics(ctx)
}

func (uc *UseCase) UpdateConfig(ctx context.Context, cfg *domain.LDAPConfig) error {
	if err := uc.settingsUseCase.UpdateLDAPConfig(ctx, cfg); err != nil {
		return fmt.Errorf("failed to update LDAP config: %w", err)
	}

	go func() { //nolint:contextcheck // background context is appropriate here
		if err := uc.ldapService.ReloadConfig(context.Background()); err != nil {
			slog.Error("Failed to reload LDAP config", "error", err)
		}
	}()

	return nil
}

func (uc *UseCase) Disable(ctx context.Context) error {
	if err := uc.settingsUseCase.UpdateLDAPConfig(ctx, &domain.LDAPConfig{Enabled: false}); err != nil {
		return fmt.Errorf("failed to update LDAP config: %w", err)
	}

	if err := uc.ldapService.ReloadConfig(ctx); err != nil {
		return fmt.Errorf("failed to reload LDAP config: %w", err)
	}

	return nil
}
