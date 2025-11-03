package rbac

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/db"
)

// Roles repository implementation.
// It implements contract.RolesRepository.
type Roles struct {
	db db.Tx
}

func NewRoles(pool *pgxpool.Pool) *Roles {
	return &Roles{db: pool}
}

func (r *Roles) List(ctx context.Context) ([]domain.Role, error) {
	exec := getExecutor(ctx, r.db)

	const query = `select * from workflows.roles order by created_at desc`

	rows, err := exec.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[roleModel])
	if err != nil {
		return nil, fmt.Errorf("collect roles: %w", err)
	}

	res := make([]domain.Role, 0, len(models))
	for _, m := range models {
		res = append(res, m.toDomain())
	}

	return res, nil
}

func (r *Roles) GetByKey(ctx context.Context, key string) (domain.Role, error) {
	exec := getExecutor(ctx, r.db)

	const query = `select * from workflows.roles where key = $1 limit 1`

	rows, err := exec.Query(ctx, query, key)
	if err != nil {
		return domain.Role{}, fmt.Errorf("get role by key: %w", err)
	}
	defer rows.Close()

	role, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[roleModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Role{}, domain.ErrEntityNotFound
		}

		return domain.Role{}, fmt.Errorf("collect role: %w", err)
	}

	return role.toDomain(), nil
}

func (r *Roles) GetByID(ctx context.Context, id domain.RoleID) (domain.Role, error) {
	exec := getExecutor(ctx, r.db)

	const query = `select * from workflows.roles where id = $1 limit 1`

	rows, err := exec.Query(ctx, query, id)
	if err != nil {
		return domain.Role{}, fmt.Errorf("get role by key: %w", err)
	}
	defer rows.Close()

	role, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[roleModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Role{}, domain.ErrEntityNotFound
		}

		return domain.Role{}, fmt.Errorf("collect role: %w", err)
	}

	return role.toDomain(), nil
}

var _ contract.RolesRepository = (*Roles)(nil)

// Permissions repository implementation.
// It implements contract.PermissionsRepository.
type Permissions struct {
	db db.Tx
}

func NewPermissions(pool *pgxpool.Pool) *Permissions {
	return &Permissions{db: pool}
}

func (r *Permissions) List(ctx context.Context) ([]domain.Permission, error) {
	exec := getExecutor(ctx, r.db)

	const query = `select * from workflows.permissions order by key`

	rows, err := exec.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[permissionModel])
	if err != nil {
		return nil, fmt.Errorf("collect permissions: %w", err)
	}

	res := make([]domain.Permission, 0, len(models))
	for _, m := range models {
		res = append(res, m.toDomain())
	}

	return res, nil
}

func (r *Permissions) ListForRole(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	exec := getExecutor(ctx, r.db)

	const query = `
		select p.* from workflows.role_permissions rp
		join workflows.permissions p on p.id = rp.permission_id
		where rp.role_id = $1
		order by p.key
	`

	rows, err := exec.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("list permissions for role: %w", err)
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[permissionModel])
	if err != nil {
		return nil, fmt.Errorf("collect role permissions: %w", err)
	}

	res := make([]domain.Permission, 0, len(models))
	for _, m := range models {
		res = append(res, m.toDomain())
	}

	return res, nil
}

func (r *Permissions) ListForAllRoles(ctx context.Context) (map[domain.Role][]domain.Permission, error) {
	exec := getExecutor(ctx, r.db)

	const query = `
		select r.id as id, r.key as key, r.name as name, r.description as description, r.created_at as created_at,
		       p.id as p_id, p.key as p_key, p.name as p_name
		from workflows.roles r
		left join workflows.role_permissions rp on rp.role_id = r.id
		left join workflows.permissions p on p.id = rp.permission_id
		order by r.created_at desc, p.key
	`

	rows, err := exec.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list permissions for all roles: %w", err)
	}
	defer rows.Close()

	type row struct {
		roleModel
		P_ID   *string `db:"p_id"`
		P_Key  *string `db:"p_key"`
		P_Name *string `db:"p_name"`
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[row])
	if err != nil {
		return nil, fmt.Errorf("collect role-permission rows: %w", err)
	}

	result := make(map[domain.Role][]domain.Permission)
	for _, it := range items {
		role := it.toDomain()
		if _, ok := result[role]; !ok {
			result[role] = []domain.Permission{}
		}
		if it.P_ID != nil {
			perm := domain.Permission{ID: domain.PermissionID(*it.P_ID), Key: domain.PermKey(*it.P_Key), Name: *it.P_Name}
			result[role] = append(result[role], perm)
		}
	}

	return result, nil
}

func (r *Permissions) RoleHasPermission(
	ctx context.Context,
	roleID string,
	key domain.PermKey,
) (bool, error) {
	exec := getExecutor(ctx, r.db)

	const query = `select exists(
select 1 from workflows.role_permissions rp
join workflows.permissions p on p.id = rp.permission_id
where rp.role_id = $1 and p.key = $2
)`

	var has bool
	if err := exec.QueryRow(ctx, query, roleID, string(key)).Scan(&has); err != nil {
		return false, fmt.Errorf("role has permission: %w", err)
	}

	return has, nil
}

var _ contract.PermissionsRepository = (*Permissions)(nil)

// Memberships repository implementation.
// It implements contract.MembershipsRepository.
type Memberships struct {
	db db.Tx
}

func NewMemberships(pool *pgxpool.Pool) *Memberships {
	return &Memberships{db: pool}
}

func (r *Memberships) GetForUserProject(
	ctx context.Context,
	userID domain.UserID,
	projectID domain.ProjectID,
) (string, error) { // roleID
	exec := getExecutor(ctx, r.db)

	const query = `select role_id from workflows.memberships where project_id = $1 and user_id = $2 limit 1`

	var roleID string
	if err := exec.QueryRow(ctx, query, projectID, userID).Scan(&roleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}

		return "", fmt.Errorf("get membership for user/project: %w", err)
	}

	return roleID, nil
}

func (r *Memberships) ListForProject(
	ctx context.Context,
	projectID domain.ProjectID,
) ([]domain.ProjectMembership, error) {
	exec := getExecutor(ctx, r.db)

	const query = `
select m.id, m.project_id, m.user_id, m.role_id, r.key as role_key, r.name as role_name, m.created_at
from workflows.memberships m
join workflows.roles r on r.id = m.role_id
where m.project_id = $1
order by m.created_at desc`

	rows, err := exec.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project memberships: %w", err)
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[membershipModel])
	if err != nil {
		return nil, fmt.Errorf("collect memberships: %w", err)
	}

	res := make([]domain.ProjectMembership, 0, len(models))
	for _, m := range models {
		res = append(res, m.toDomain())
	}

	return res, nil
}

func (r *Memberships) Create(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
	roleID domain.RoleID,
) (domain.ProjectMembership, error) {
	exec := getExecutor(ctx, r.db)

	const query = `
with ins as (
	insert into workflows.memberships (project_id, user_id, role_id)
	values ($1, $2, $3)
	returning id, project_id, user_id, role_id, created_at
)
select ins.id, ins.project_id, ins.user_id, ins.role_id, r.key as role_key, r.name as role_name, ins.created_at
from ins join workflows.roles r on r.id = ins.role_id`

	row := exec.QueryRow(ctx, query, projectID, userID, roleID)
	var model membershipModel
	if err := row.Scan(
		&model.ID,
		&model.ProjectID,
		&model.UserID,
		&model.RoleID,
		&model.RoleKey,
		&model.RoleName,
		&model.CreatedAt,
	); err != nil {
		return domain.ProjectMembership{}, fmt.Errorf("insert membership: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Memberships) Get(
	ctx context.Context,
	projectID domain.ProjectID,
	membershipID domain.MembershipID,
) (domain.ProjectMembership, error) {
	exec := getExecutor(ctx, r.db)

	const query = `
select m.id, m.project_id, m.user_id, m.role_id, r.key as role_key, r.name as role_name, m.created_at
from workflows.memberships m
join workflows.roles r on r.id = m.role_id
where m.project_id = $1 and m.id = $2
limit 1`

	row := exec.QueryRow(ctx, query, projectID, membershipID)
	var model membershipModel
	if err := row.Scan(
		&model.ID,
		&model.ProjectID,
		&model.UserID,
		&model.RoleID,
		&model.RoleKey,
		&model.RoleName,
		&model.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectMembership{}, domain.ErrEntityNotFound
		}

		return domain.ProjectMembership{}, fmt.Errorf("get membership: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Memberships) Update(
	ctx context.Context,
	projectID domain.ProjectID,
	membershipID domain.MembershipID,
	roleID domain.RoleID,
) (domain.ProjectMembership, error) {
	exec := getExecutor(ctx, r.db)

	const query = `
with upd as (
	update workflows.memberships set role_id = $3, updated_at = now()
	where project_id = $1 and id = $2
	returning id, project_id, user_id, role_id, created_at
)
select upd.id, upd.project_id, upd.user_id, upd.role_id, r.key as role_key, r.name as role_name, upd.created_at
from upd join workflows.roles r on r.id = upd.role_id`

	row := exec.QueryRow(ctx, query, projectID, membershipID, roleID)
	var model membershipModel
	if err := row.Scan(
		&model.ID,
		&model.ProjectID,
		&model.UserID,
		&model.RoleID,
		&model.RoleKey,
		&model.RoleName,
		&model.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectMembership{}, domain.ErrEntityNotFound
		}

		return domain.ProjectMembership{}, fmt.Errorf("update membership: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Memberships) Delete(ctx context.Context, projectID domain.ProjectID, membershipID domain.MembershipID) error {
	exec := getExecutor(ctx, r.db)

	const query = `delete from workflows.memberships where project_id = $1 and id = $2`

	ct, err := exec.Exec(ctx, query, projectID, membershipID)
	if err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
	}

	return nil
}

var _ contract.MembershipsRepository = (*Memberships)(nil)

// helper to get tx from context
//
//nolint:ireturn // internal helper
func getExecutor(ctx context.Context, fallback db.Tx) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return fallback
}
