package permissions

import (
	"context"
	"log/slog"

	etx "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

// Service handles permission checks for various operations.
type Service struct {
	projects contract.ProjectsRepository
	roles    contract.RolesRepository
	perms    contract.PermissionsRepository
	member   contract.MembershipsRepository
}

// New creates a new permissions service.
func New(
	projects contract.ProjectsRepository,
	roles contract.RolesRepository,
	perms contract.PermissionsRepository,
	member contract.MembershipsRepository,
) *Service {
	return &Service{projects: projects, roles: roles, perms: perms, member: member}
}

func (s *Service) isSuper(ctx context.Context) bool { return etx.IsSuper(ctx) }

// HasGlobalPermission checks global (non-project) permissions.
// For now, only superuser has global permissions.
func (s *Service) HasGlobalPermission(
	ctx context.Context,
	permKey domain.PermKey,
) (bool, error) {
	_ = permKey // reserved for future global-permissions storage

	return s.isSuper(ctx), nil
}

// HasProjectPermission checks if the user has a specific permission in the scope of the project.
func (s *Service) HasProjectPermission(
	ctx context.Context,
	projectID domain.ProjectID,
	permKey domain.PermKey,
) (bool, error) {
	if s.isSuper(ctx) {
		return true, nil
	}

	userID := etx.UserID(ctx)
	if userID == 0 {
		slog.Debug("HasProjectPermission: userID is 0", "project_id", projectID, "permission", permKey)
		return false, domain.ErrUserNotFound
	}

	// Verify the project exists (preserve current behavior and error mapping)
	if _, err := s.projects.GetByID(ctx, projectID); err != nil {
		slog.Debug("HasProjectPermission: project not found", "error", err, "project_id", projectID, "user_id", userID, "permission", permKey)
		return false, err
	}

	roleID, err := s.member.GetForUserProject(ctx, userID, projectID)
	if err != nil {
		slog.Debug("HasProjectPermission: failed to get membership", "error", err, "project_id", projectID, "user_id", userID, "permission", permKey)
		return false, err
	}
	if roleID == "" {
		slog.Debug("HasProjectPermission: no membership found", "project_id", projectID, "user_id", userID, "permission", permKey)
		return false, nil
	}

	hasPerm, err := s.perms.RoleHasPermission(ctx, roleID, permKey)
	if err != nil {
		slog.Error("HasProjectPermission: failed to check role permission",
			"error", err,
			"project_id", projectID,
			"user_id", userID,
			"role_id", roleID,
			"permission", permKey,
		)
		return false, err
	}

	slog.Debug("HasProjectPermission: result",
		"project_id", projectID,
		"user_id", userID,
		"role_id", roleID,
		"permission", permKey,
		"has_permission", hasPerm,
	)

	return hasPerm, nil
}

// CanAccessProject checks if a user can access a project.
func (s *Service) CanAccessProject(ctx context.Context, projectID domain.ProjectID) error {
	ok, err := s.HasProjectPermission(ctx, projectID, domain.PermProjectView)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrPermissionDenied
	}

	return nil
}

// CanViewProject checks if a user can view a project (basic access for all project members).
// This is more permissive than CanAccessProject and allows all project members to view project data.
func (s *Service) CanViewProject(ctx context.Context, projectID domain.ProjectID) error {
	// Superusers can always view projects
	if s.isSuper(ctx) {
		return nil
	}

	userID := etx.UserID(ctx)
	if userID == 0 {
		return domain.ErrUserNotFound
	}

	// Check if a user has any membership in the project (any role)
	roleID, err := s.member.GetForUserProject(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if roleID == "" {
		return domain.ErrPermissionDenied
	}

	return nil
}

// CanManageProject checks if a user can manage a project (create, update, delete).
func (s *Service) CanManageProject(ctx context.Context, projectID domain.ProjectID) error {
	ok, err := s.HasProjectPermission(ctx, projectID, domain.PermProjectManage)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrPermissionDenied
	}

	return nil
}

// CanViewAudit checks if a user can view audit logs.
func (s *Service) CanViewAudit(ctx context.Context, projectID domain.ProjectID) error {
	ok, err := s.HasProjectPermission(ctx, projectID, domain.PermAuditView)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrPermissionDenied
	}

	return nil
}

// CanManageMembership checks if a user can manage project memberships.
func (s *Service) CanManageMembership(ctx context.Context, projectID domain.ProjectID) error {
	ok, err := s.HasProjectPermission(ctx, projectID, domain.PermMembershipManage)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrPermissionDenied
	}

	return nil
}

// CanCreateWorkflow checks if a user can create workflows in a project.
func (s *Service) CanCreateWorkflow(ctx context.Context, projectID domain.ProjectID) error {
	ok, err := s.HasProjectPermission(ctx, projectID, domain.PermWorkflowCreate)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrPermissionDenied
	}

	return nil
}

// GetAccessibleProjects returns all projects that a user can access.
func (s *Service) GetAccessibleProjects(
	ctx context.Context,
	projects []domain.Project,
) ([]domain.Project, error) {
	if s.isSuper(ctx) {
		return projects, nil
	}

	out := make([]domain.Project, 0, len(projects))

	for i := range projects {
		project := projects[i]

		ok, err := s.HasProjectPermission(ctx, project.ID, domain.PermProjectView)
		if err != nil {
			return nil, err
		}

		if ok {
			out = append(out, project)
		}
	}

	return out, nil
}

// GetMyProjectPermissions returns permissions for projects where the user has a membership.
func (s *Service) GetMyProjectPermissions(
	ctx context.Context,
) (map[domain.ProjectID][]domain.PermKey, error) {
	userID := etx.UserID(ctx)
	if userID == 0 {
		return nil, domain.ErrUserNotFound
	}

	all, err := s.projects.List(ctx)
	if err != nil {
		return nil, err
	}

	// Define the set of permission keys we expose via this endpoint
	permKeys := []domain.PermKey{
		domain.PermProjectView,
		domain.PermProjectManage,
		domain.PermProjectCreate,
		domain.PermWorkflowCreate,
		domain.PermAuditView,
		domain.PermMembershipManage,
	}

	result := make(map[domain.ProjectID][]domain.PermKey)

	for i := range all {
		project := all[i]

		// Check membership directly, do not use superuser bypass here
		roleID, mErr := s.member.GetForUserProject(ctx, userID, project.ID)
		if mErr != nil {
			// Continue to next project if membership not found (expected for projects user doesn't belong to)
			continue
		}

		if roleID == "" {
			continue // no membership — skip this project
		}

		// Collect granted permissions for the role
		var granted []domain.PermKey

		for _, key := range permKeys {
			has, perr := s.perms.RoleHasPermission(ctx, roleID, key)
			if perr != nil {
				slog.Error("Failed to check role permission",
					"error", perr,
					"role_id", roleID,
					"permission", key,
					"project_id", project.ID,
					"user_id", userID,
				)
				return nil, perr
			}

			if has {
				slog.Debug("Permission granted",
					"role_id", roleID,
					"permission", key,
					"project_id", project.ID,
					"user_id", userID,
				)
				granted = append(granted, key)
			}
		}

		if len(granted) > 0 {
			slog.Debug("Project permissions found",
				"project_id", project.ID,
				"user_id", userID,
				"role_id", roleID,
				"permissions_count", len(granted),
			)
			result[project.ID] = granted
		} else {
			slog.Debug("No permissions found for project",
				"project_id", project.ID,
				"user_id", userID,
				"role_id", roleID,
			)
		}
	}

	return result, nil
}

func (s *Service) GetMyProjectRoles(ctx context.Context) (map[domain.ProjectID]domain.Role, error) {
	userID := etx.UserID(ctx)
	if userID == 0 {
		return nil, domain.ErrUserNotFound
	}

	all, err := s.projects.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[domain.ProjectID]domain.Role)

	for i := range all {
		project := all[i]

		// Check membership directly, do not use superuser bypass here
		roleID, err := s.member.GetForUserProject(ctx, userID, project.ID)
		if err != nil {
			return nil, err
		}

		if roleID == "" {
			continue // no membership — skip this project
		}

		role, err := s.roles.GetByID(ctx, domain.RoleID(roleID))
		if err != nil {
			return nil, err
		}

		result[project.ID] = role
	}

	return result, nil
}
