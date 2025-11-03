package ldapsynclogs

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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

//nolint:lll // it's insert query
func (r *Repository) Create(ctx context.Context, log domain.LDAPSyncLog) (domain.LDAPSyncLog, error) {
	executor := r.getExecutor(ctx)

	const query = `
INSERT INTO  workflows_manager.ldap_sync_logs (timestamp, level, message, username, details, sync_session_id, stack_trace, ldap_error_code, ldap_error_message)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, timestamp, level, message, username, details, sync_session_id, stack_trace, ldap_error_code, ldap_error_message`

	var model ldapSyncLogModel

	err := executor.QueryRow(ctx, query,
		log.Timestamp,
		log.Level,
		log.Message,
		log.Username,
		log.Details,
		log.SyncSessionID,
		log.StackTrace,
		log.LDAPErrorCode,
		log.LDAPErrorMessage,
	).Scan(
		&model.ID,
		&model.Timestamp,
		&model.Level,
		&model.Message,
		&model.Username,
		&model.Details,
		&model.SyncSessionID,
		&model.StackTrace,
		&model.LDAPErrorCode,
		&model.LDAPErrorMessage,
	)
	if err != nil {
		return domain.LDAPSyncLog{}, fmt.Errorf("insert ldap sync log: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) GetByID(ctx context.Context, id uint) (domain.LDAPSyncLog, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM  workflows_manager.ldap_sync_logs WHERE id = $1 LIMIT 1`

	rows, err := executor.Query(ctx, query, id)
	if err != nil {
		return domain.LDAPSyncLog{}, fmt.Errorf("query ldap sync log by ID: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ldapSyncLogModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LDAPSyncLog{}, domain.ErrEntityNotFound
		}

		return domain.LDAPSyncLog{}, fmt.Errorf("collect ldap sync log: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) List(ctx context.Context, filter domain.LDAPSyncLogFilter) (domain.LDAPSyncLogsResult, error) {
	executor := r.getExecutor(ctx)

	const (
		tableName    = " workflows_manager.ldap_sync_logs"
		defaultLimit = 100
	)

	builder := sq.
		Select("*").
		From(tableName).
		OrderBy("timestamp DESC").
		PlaceholderFormat(sq.Dollar)

	countBuilder := sq.
		Select("COUNT(*)").
		From(tableName).
		PlaceholderFormat(sq.Dollar)

	applyFilters := func(builder sq.SelectBuilder) sq.SelectBuilder {
		if filter.Level != nil {
			builder = builder.Where(sq.Eq{"level": *filter.Level})
		}

		if filter.SyncID != nil {
			builder = builder.Where(sq.Eq{"sync_session_id": *filter.SyncID})
		}

		if filter.Username != nil {
			builder = builder.Where(sq.Eq{"username": *filter.Username})
		}

		if filter.From != nil {
			builder = builder.Where(sq.GtOrEq{"timestamp": *filter.From})
		}

		if filter.To != nil {
			builder = builder.Where(sq.LtOrEq{"timestamp": *filter.To})
		}

		return builder
	}

	builder = applyFilters(builder)
	countBuilder = applyFilters(countBuilder)

	limit := defaultLimit
	if filter.Limit != nil && *filter.Limit > 0 {
		limit = *filter.Limit
	}

	builder = builder.Limit(uint64(limit)) //nolint:gosec // it's ok

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return domain.LDAPSyncLogsResult{}, fmt.Errorf("build select query: %w", err)
	}

	rows, err := executor.Query(ctx, sqlStr, args...)
	if err != nil {
		return domain.LDAPSyncLogsResult{}, fmt.Errorf("execute select query: %w", err)
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[ldapSyncLogModel])
	if err != nil {
		return domain.LDAPSyncLogsResult{}, fmt.Errorf("collect rows: %w", err)
	}

	logs := make([]domain.LDAPSyncLog, 0, len(models))
	for _, m := range models {
		logs = append(logs, m.toDomain())
	}

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return domain.LDAPSyncLogsResult{}, fmt.Errorf("build count query: %w", err)
	}

	var total int
	if err := executor.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return domain.LDAPSyncLogsResult{}, fmt.Errorf("execute count query: %w", err)
	}

	return domain.LDAPSyncLogsResult{
		Logs:  logs,
		Total: total,
	}, nil
}

func (r *Repository) DeleteBySyncID(ctx context.Context, syncSessionID string) error {
	executor := r.getExecutor(ctx)

	const query = `DELETE FROM  workflows_manager.ldap_sync_logs WHERE sync_session_id = $1`

	_, err := executor.Exec(ctx, query, syncSessionID)
	if err != nil {
		return fmt.Errorf("delete ldap sync logs by sync session ID: %w", err)
	}

	return nil
}

//nolint:ireturn // it's ok here
func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
