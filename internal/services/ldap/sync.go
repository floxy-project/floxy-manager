package ldap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

var (
	ErrSyncInProgress = errors.New("sync already in progress")
	ErrSyncNotRunning = errors.New("no sync is currently running")
)

// StartManualSync starts a manual LDAP synchronization of the specified type.
//
//nolint:gocyclo,gocognit,contextcheck // appropriate
func (s *Service) StartManualSync(_ context.Context, syncID string, stopped chan struct{}) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	// Check if sync is already running
	if s.syncProgress.IsRunning {
		return ErrSyncInProgress
	}

	// Check if LDAP is enabled and configured
	if !s.enabled || s.client == nil {
		return ErrLDAPNotConfigured
	}

	// Initialize sync progress
	s.syncProgress = domain.LDAPSyncProgress{
		IsRunning:      true,
		Progress:       0,
		CurrentStep:    "Starting synchronization",
		ProcessedItems: 0,
		TotalItems:     0,
		StartTime:      time.Now(),
		SyncID:         syncID,
	}
	s.syncStatus = domain.LDAPSyncStatus{
		IsRunning:    true,
		LastSyncTime: time.Now(),
		TotalUsers:   0,
		SyncedUsers:  0,
		Errors:       0,
		Warnings:     0,
	}

	// Create a cancellable context for this sync
	syncCtx, cancelFunc := context.WithCancel(s.ctx)
	s.syncCancelCtx = cancelFunc

	// Start sync in the background
	go func() {
		defer func() {
			if stopped != nil {
				close(stopped)
			}
		}()

		time.Sleep(2 * time.Second)

		startTime := time.Now()

		var syncErr error

		var totalUsers, syncedUsers, errs, warnings int

		defer func() {
			duration := time.Since(startTime).Truncate(time.Second)

			// Log the sync result
			level := "info"
			message := fmt.Sprintf("LDAP sync completed successfully. Duration: %s", duration)
			details := fmt.Sprintf("Total users: %d, Synced users: %d, Errors: %d, Warnings: %d",
				totalUsers, syncedUsers, errs, warnings)

			if syncErr != nil {
				level = "error"
				message = fmt.Sprintf("LDAP sync failed. Duration: %s, Error: %v", duration, syncErr)
				details = fmt.Sprintf("Total users: %d, Synced users: %d, Errors: %d, Warnings: %d",
					totalUsers, syncedUsers, errs, warnings)
			}

			// Create log entry
			log := domain.LDAPSyncLog{
				Timestamp:        time.Now(),
				Level:            level,
				Message:          message,
				Username:         nil,
				Details:          &details,
				SyncSessionID:    syncID,
				StackTrace:       nil,
				LDAPErrorCode:    nil,
				LDAPErrorMessage: nil,
			}

			// Write log to a database
			if _, logErr := s.ldapSyncLogsRepo.Create(syncCtx, log); logErr != nil {
				slog.Error("Failed to write LDAP sync log", "error", logErr, "syncID", syncID)
			}

			// Update sync stats in a database
			endTime := time.Now()
			durationStr := duration.String()
			status := "completed"

			var errorMessage *string

			if syncErr != nil {
				status = "failed"
				errMsg := syncErr.Error()
				errorMessage = &errMsg
			}

			stats := domain.LDAPSyncStats{
				SyncSessionID: syncID,
				StartTime:     startTime,
				EndTime:       &endTime,
				Duration:      &durationStr,
				TotalUsers:    totalUsers,
				SyncedUsers:   syncedUsers,
				Errors:        errs,
				Warnings:      warnings,
				Status:        status,
				ErrorMessage:  errorMessage,
			}

			// Try to update the existing stats record
			if updateErr := s.ldapSyncStatsRepo.Update(syncCtx, stats); updateErr != nil {
				slog.Error("Failed to update LDAP sync stats", "error", updateErr, "syncID", syncID)
			}

			// Update the sync status when complete
			s.syncMutex.Lock()
			s.syncStatus = domain.LDAPSyncStatus{
				IsRunning:        false,
				LastSyncTime:     startTime,
				TotalUsers:       totalUsers,
				SyncedUsers:      syncedUsers,
				Errors:           errs,
				Warnings:         warnings,
				LastSyncDuration: duration.String(),
			}
			s.syncStatus.IsRunning = false
			s.syncProgress.IsRunning = false
			s.syncProgress.Progress = 100
			s.syncProgress.CurrentStep = "Completed"
			s.syncMutex.Unlock()

			if syncErr != nil {
				slog.Error("LDAP sync failed", "error", syncErr, "duration", duration)
			} else {
				slog.Info("LDAP sync completed", "duration", duration)
			}
		}()

		// Get all users from LDAP first to set the total count
		users, err := s.client.GetAllUsers(syncCtx)
		if err != nil {
			syncErr = fmt.Errorf("failed to get users from LDAP: %w", err)

			return
		}

		totalUsers = len(users)

		s.syncMutex.Lock()
		s.syncProgress.TotalItems = totalUsers
		s.syncProgress.CurrentStep = "Syncing users"
		s.syncMutex.Unlock()

		// Update initial stats
		s.updateSyncStats(syncCtx, syncID, map[string]any{
			"total_users": totalUsers,
		})

		// Sync each user
		for i, username := range users {
			select {
			case <-syncCtx.Done():
				return
			default:
				time.Sleep(time.Millisecond * 100)

				if err := s.syncUser(syncCtx, username); err != nil {
					// Log error but continue with other users
					slog.Error("failed to sync user", "username", username, "error", err)

					errs++
				} else {
					syncedUsers++
				}

				// Update progress
				s.syncMutex.Lock()
				s.syncProgress.ProcessedItems = i + 1
				s.syncProgress.Progress = float64(i+1) * 100 / float64(totalUsers)
				s.syncMutex.Unlock()

				// Update stats every 10 users or at the end
				if (i+1)%10 == 0 || i == len(users)-1 {
					s.updateSyncStats(syncCtx, syncID, map[string]any{
						"synced_users": syncedUsers,
						"errors":       errs,
					})
				}
			}
		}
	}()

	return nil
}

// CancelSync cancels an ongoing synchronization.
func (s *Service) CancelSync(ctx context.Context) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	if !s.syncProgress.IsRunning {
		return ErrSyncNotRunning
	}

	if s.syncCancelCtx != nil {
		s.syncCancelCtx()
	}

	s.syncProgress.IsRunning = false
	s.syncProgress.CurrentStep = "Cancelled"

	// Update sync stats to mark as cancelled
	endTime := time.Now()
	duration := time.Since(s.syncProgress.StartTime)
	durationStr := duration.String()
	status := "cancelled"

	stats := domain.LDAPSyncStats{
		SyncSessionID: s.syncProgress.SyncID,
		EndTime:       &endTime,
		Duration:      &durationStr,
		Status:        status,
	}

	// Try to update the stats
	if err := s.ldapSyncStatsRepo.Update(ctx, stats); err != nil {
		slog.Error("Failed to update sync stats on cancel", "error", err, "syncID", s.syncProgress.SyncID)
	}

	return nil
}

// GetSyncStatus returns the current status of LDAP synchronization.
func (s *Service) GetSyncStatus() domain.LDAPSyncStatus {
	s.syncMutex.RLock()
	defer s.syncMutex.RUnlock()

	return s.syncStatus
}

// GetSyncProgress returns the progress of an ongoing LDAP synchronization.
func (s *Service) GetSyncProgress() domain.LDAPSyncProgress {
	s.syncMutex.RLock()
	defer s.syncMutex.RUnlock()

	return s.syncProgress
}

// updateSyncStats updates the sync statistics in the database.
func (s *Service) updateSyncStats(ctx context.Context, syncID string, updates map[string]any) {
	stats, err := s.ldapSyncStatsRepo.GetBySyncSessionID(ctx, syncID)
	if err != nil {
		slog.Error("Failed to get sync stats", "error", err, "syncID", syncID)

		return
	}

	// Apply updates
	if totalUsers, ok := updates["total_users"].(int); ok {
		stats.TotalUsers = totalUsers
	}

	if syncedUsers, ok := updates["synced_users"].(int); ok {
		stats.SyncedUsers = syncedUsers
	}

	if errs, ok := updates["errors"].(int); ok {
		stats.Errors = errs
	}

	if warnings, ok := updates["warnings"].(int); ok {
		stats.Warnings = warnings
	}

	// Try to update the stats
	if err := s.ldapSyncStatsRepo.Update(ctx, stats); err != nil {
		slog.Error("Failed to update sync stats", "error", err, "syncID", syncID)
	}
}
