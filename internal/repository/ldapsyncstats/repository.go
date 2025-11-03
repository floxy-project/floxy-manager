package ldapsyncstats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/db"
)

type Repository struct {
	db db.Tx
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		db: pool,
	}
}

func (r *Repository) Create(ctx context.Context, stats domain.LDAPSyncStats) (domain.LDAPSyncStats, error) {
	executor := r.getExecutor(ctx)

	const query = `
INSERT INTO  workflows_manager.ldap_sync_stats (sync_session_id, start_time, end_time, duration, total_users,
                             synced_users, errors, warnings, status, error_message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, sync_session_id, start_time, end_time, duration, total_users, 
    synced_users, errors, warnings, status, error_message`

	var model ldapSyncStatsModel

	err := executor.QueryRow(ctx, query,
		stats.SyncSessionID,
		stats.StartTime,
		stats.EndTime,
		stats.Duration,
		stats.TotalUsers,
		stats.SyncedUsers,
		stats.Errors,
		stats.Warnings,
		stats.Status,
		stats.ErrorMessage,
	).Scan(
		&model.ID,
		&model.SyncSessionID,
		&model.StartTime,
		&model.EndTime,
		&model.Duration,
		&model.TotalUsers,
		&model.SyncedUsers,
		&model.Errors,
		&model.Warnings,
		&model.Status,
		&model.ErrorMessage,
	)
	if err != nil {
		return domain.LDAPSyncStats{}, fmt.Errorf("insert ldap sync stats: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) GetBySyncSessionID(ctx context.Context, syncSessionID string) (domain.LDAPSyncStats, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM  workflows_manager.ldap_sync_stats WHERE sync_session_id = $1 LIMIT 1`

	rows, err := executor.Query(ctx, query, syncSessionID)
	if err != nil {
		return domain.LDAPSyncStats{}, fmt.Errorf("query ldap sync stats by sync session ID: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ldapSyncStatsModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LDAPSyncStats{}, domain.ErrEntityNotFound
		}

		return domain.LDAPSyncStats{}, fmt.Errorf("collect ldap sync stats: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) Update(ctx context.Context, stats domain.LDAPSyncStats) error {
	executor := r.getExecutor(ctx)

	const query = `
UPDATE  workflows_manager.ldap_sync_stats
SET start_time = $1, end_time = $2, duration = $3, total_users = $4,
    synced_users = $5, errors = $6, warnings = $7, status = $8, error_message = $9
WHERE sync_session_id = $10`

	tag, err := executor.Exec(ctx, query,
		stats.StartTime,
		stats.EndTime,
		stats.Duration,
		stats.TotalUsers,
		stats.SyncedUsers,
		stats.Errors,
		stats.Warnings,
		stats.Status,
		stats.ErrorMessage,
		stats.SyncSessionID,
	)
	if err != nil {
		return fmt.Errorf("update ldap sync stats: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, limit int) ([]domain.LDAPSyncStats, error) {
	executor := r.getExecutor(ctx)

	if limit <= 0 {
		limit = 50
	}

	if limit > 1000 {
		limit = 1000
	}

	const query = `SELECT * FROM  workflows_manager.ldap_sync_stats ORDER BY start_time DESC LIMIT $1`

	rows, err := executor.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query ldap sync stats: %w", err)
	}

	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[ldapSyncStatsModel])
	if err != nil {
		return nil, fmt.Errorf("collect ldap sync stats: %w", err)
	}

	stats := make([]domain.LDAPSyncStats, 0, len(models))
	for _, model := range models {
		stats = append(stats, model.toDomain())
	}

	return stats, nil
}

func (r *Repository) ListCompleted(ctx context.Context, limit int) ([]domain.LDAPSyncStats, error) {
	executor := r.getExecutor(ctx)

	if limit <= 0 {
		limit = 50
	}

	if limit > 1000 {
		limit = 1000
	}

	const query = `SELECT * FROM  workflows_manager.ldap_sync_stats WHERE status = 'completed' ORDER BY start_time DESC LIMIT $1`

	rows, err := executor.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query ldap sync stats: %w", err)
	}

	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[ldapSyncStatsModel])
	if err != nil {
		return nil, fmt.Errorf("collect ldap sync stats: %w", err)
	}

	stats := make([]domain.LDAPSyncStats, 0, len(models))
	for _, model := range models {
		stats = append(stats, model.toDomain())
	}

	return stats, nil
}

func (r *Repository) GetStatistics(ctx context.Context) (domain.LDAPStatistics, error) {
	executor := r.getExecutor(ctx)

	// Get basic counts
	const basicQuery = `
SELECT 
    COUNT(DISTINCT u.id) as local_users,
    COUNT(DISTINCT CASE WHEN u.is_active = true THEN u.id END) as active_users,
    COUNT(DISTINCT CASE WHEN u.is_active = false THEN u.id END) as inactive_users
FROM  workflows_manager.users u`

	var localUsers, activeUsers, inactiveUsers int

	err := executor.QueryRow(ctx, basicQuery).Scan(&localUsers, &activeUsers, &inactiveUsers)
	if err != nil {
		return domain.LDAPStatistics{}, fmt.Errorf("query basic statistics: %w", err)
	}

	// Get sync history (last 30 days)
	const historyQuery = `
SELECT 
    DATE(start_time) as date,
    SUM(synced_users) as users_synced,
    SUM(errors) as errors,
    AVG(EXTRACT(EPOCH FROM (end_time - start_time))/60) as duration_minutes
FROM  workflows_manager.ldap_sync_stats 
WHERE start_time >= $1 AND status = 'completed'
GROUP BY DATE(start_time)
ORDER BY date DESC
LIMIT 30`

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	rows, err := executor.Query(ctx, historyQuery, thirtyDaysAgo)
	if err != nil {
		return domain.LDAPStatistics{}, fmt.Errorf("query sync history: %w", err)
	}

	defer rows.Close()

	var history []domain.LDAPStatisticsSyncHistory

	for rows.Next() {
		var syncHistory domain.LDAPStatisticsSyncHistory

		var dateStr string

		err := rows.Scan(&dateStr, &syncHistory.UsersSynced, &syncHistory.Errors, &syncHistory.DurationMinutes)
		if err != nil {
			return domain.LDAPStatistics{}, fmt.Errorf("scan sync history: %w", err)
		}

		syncHistory.Date, _ = time.Parse("2006-01-02", dateStr)
		history = append(history, syncHistory)
	}

	// Calculate success rate
	const successQuery = `
SELECT 
    COUNT(*) as total_syncs,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_syncs
FROM  workflows_manager.ldap_sync_stats 
WHERE start_time >= $1`

	rows, err = executor.Query(ctx, successQuery, thirtyDaysAgo)
	if err != nil {
		return domain.LDAPStatistics{}, fmt.Errorf("query success rate: %w", err)
	}
	defer rows.Close()

	var totalSyncs, successfulSyncs int
	if rows.Next() {
		err = rows.Scan(&totalSyncs, &successfulSyncs)
		if err != nil {
			return domain.LDAPStatistics{}, fmt.Errorf("scan success rate: %w", err)
		}
	}

	var successRate float32
	if totalSyncs > 0 {
		successRate = float32(successfulSyncs) / float32(totalSyncs) * 100
	}

	// Get LDAP counts from latest successful sync
	const ldapQuery = `
SELECT total_users
FROM  workflows_manager.ldap_sync_stats 
WHERE status = 'completed'
ORDER BY start_time DESC
LIMIT 1`

	var ldapUsers int

	err = executor.QueryRow(ctx, ldapQuery).Scan(&ldapUsers)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.LDAPStatistics{}, fmt.Errorf("query LDAP counts: %w", err)
	}

	return domain.LDAPStatistics{
		LDAPUsers:       ldapUsers,
		LocalUsers:      localUsers,
		ActiveUsers:     activeUsers,
		InactiveUsers:   inactiveUsers,
		SyncHistory:     history,
		SyncSuccessRate: successRate,
	}, nil
}

//nolint:ireturn // it's ok here
func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
