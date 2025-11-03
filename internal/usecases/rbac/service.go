package rbac

import (
	"context"
	"fmt"

	appctx "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/internal/repository/membershipaudit"
	"github.com/rom8726/floxy-manager/pkg/db"
)

var _ contract.MembershipsUseCase = (*Service)(nil)

type Service struct {
	projectsRepo    contract.ProjectsRepository
	rolesRepo       contract.RolesRepository
	permsRepo       contract.PermissionsRepository
	membershipsRepo contract.MembershipsRepository
	tx              db.TxManager
}

func New(
	projectsRepo contract.ProjectsRepository,
	rolesRepo contract.RolesRepository,
	permsRepo contract.PermissionsRepository,
	membershipsRepo contract.MembershipsRepository,
	tx db.TxManager,
) *Service {
	return &Service{
		projectsRepo:    projectsRepo,
		rolesRepo:       rolesRepo,
		permsRepo:       permsRepo,
		membershipsRepo: membershipsRepo,
		tx:              tx,
	}
}

// Roles & permissions

func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) {
	return s.rolesRepo.List(ctx)
}

func (s *Service) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.permsRepo.List(ctx)
}

func (s *Service) GetRolePermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	return s.permsRepo.ListForRole(ctx, roleID)
}

func (s *Service) ListRolePermissions(ctx context.Context) (map[domain.Role][]domain.Permission, error) {
	return s.permsRepo.ListForAllRoles(ctx)
}

// Memberships

func (s *Service) ListProjectMemberships(
	ctx context.Context,
	projectID domain.ProjectID,
) ([]domain.ProjectMembership, error) {
	return s.membershipsRepo.ListForProject(ctx, projectID)
}

func (s *Service) GetProjectMembership(
	ctx context.Context,
	projectID domain.ProjectID,
	membershipID domain.MembershipID,
) (domain.ProjectMembership, error) {
	return s.membershipsRepo.Get(ctx, projectID, membershipID)
}

func (s *Service) CreateProjectMembership(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
	roleID domain.RoleID,
) (domain.ProjectMembership, error) {
	//project, err := s.projectsRepo.GetByID(ctx, projectID)
	//if err != nil {
	//	return domain.ProjectMembership{}, fmt.Errorf("get project: %w", err)
	//}
	//
	//role, err := s.rolesRepo.GetByID(ctx, roleID)
	//if err != nil {
	//	return domain.ProjectMembership{}, fmt.Errorf("get role: %w", err)
	//}

	var created domain.ProjectMembership
	if err := s.tx.ReadCommitted(ctx, func(ctx context.Context) error {
		membership, err := s.membershipsRepo.Create(ctx, projectID, userID, roleID)
		if err != nil {
			return err
		}
		created = membership

		actorID := int(appctx.UserID(ctx))
		exec := db.TxFromContext(ctx)
		err = membershipaudit.Write(ctx, exec,
			string(membership.ID),
			actorID,
			"create",
			nil,
			membership,
		)
		if err != nil {
			return fmt.Errorf("write membership audit: %w", err)
		}

		//content := domain.UserNotificationContent{
		//	UserAddedToProject: &domain.UserAddedToProjectContent{
		//		ProjectName: project.Name,
		//		RoleName:    role.Name,
		//		ByUser:      appctx.Username(ctx),
		//	},
		//}
		//err = s.userNotificationsUC.CreateNotification(
		//	ctx,
		//	userID,
		//	domain.UserNotificationTypeProjectAdded,
		//	content,
		//)
		//if err != nil {
		//	return fmt.Errorf("create user notification: %w", err)
		//}

		return nil
	}); err != nil {
		return domain.ProjectMembership{}, err
	}

	return created, nil
}

func (s *Service) UpdateProjectMembership(
	ctx context.Context,
	projectID domain.ProjectID,
	membershipID domain.MembershipID,
	roleID domain.RoleID,
) (domain.ProjectMembership, error) {
	//project, err := s.projectsRepo.GetByID(ctx, projectID)
	//if err != nil {
	//	return domain.ProjectMembership{}, fmt.Errorf("get project: %w", err)
	//}
	//
	//roleNew, err := s.rolesRepo.GetByID(ctx, roleID)
	//if err != nil {
	//	return domain.ProjectMembership{}, fmt.Errorf("get new role: %w", err)
	//}

	//membership, err := s.membershipsRepo.Get(ctx, projectID, membershipID)
	//if err != nil {
	//	return domain.ProjectMembership{}, fmt.Errorf("get membership: %w", err)
	//}

	//userID := membership.UserID
	//roleOld, err := s.rolesRepo.GetByID(ctx, membership.RoleID)
	//if err != nil {
	//	return domain.ProjectMembership{}, fmt.Errorf("get old role: %w", err)
	//}

	var updated domain.ProjectMembership
	if err := s.tx.ReadCommitted(ctx, func(ctx context.Context) error {
		old, err := s.membershipsRepo.Get(ctx, projectID, membershipID)
		if err != nil {
			return err
		}

		m, err := s.membershipsRepo.Update(ctx, projectID, membershipID, roleID)
		if err != nil {
			return err
		}
		updated = m

		actorID := int(appctx.UserID(ctx))
		exec := db.TxFromContext(ctx)
		if err := membershipaudit.Write(ctx, exec, string(m.ID), actorID, "update", old, m); err != nil {
			return fmt.Errorf("write membership audit: %w", err)
		}
		//
		//content := domain.UserNotificationContent{
		//	UserRoleChanged: &domain.UserRoleChangedContent{
		//		ProjectName: project.Name,
		//		RoleNameOld: roleOld.Name,
		//		RoleNameNew: roleNew.Name,
		//		ByUser:      appctx.Username(ctx),
		//	},
		//}
		//err = s.userNotificationsUC.CreateNotification(
		//	ctx,
		//	userID,
		//	domain.UserNotificationTypeRoleChanged,
		//	content,
		//)
		//if err != nil {
		//	return fmt.Errorf("create user notification: %w", err)
		//}

		return nil
	}); err != nil {
		return domain.ProjectMembership{}, err
	}

	return updated, nil
}

func (s *Service) DeleteProjectMembership(
	ctx context.Context,
	projectID domain.ProjectID,
	membershipID domain.MembershipID,
) error {
	//project, err := s.projectsRepo.GetByID(ctx, projectID)
	//if err != nil {
	//	return fmt.Errorf("get project: %w", err)
	//}
	//
	//membership, err := s.membershipsRepo.Get(ctx, projectID, membershipID)
	//if err != nil {
	//	return fmt.Errorf("get membership: %w", err)
	//}

	return s.tx.ReadCommitted(ctx, func(ctx context.Context) error {
		old, err := s.membershipsRepo.Get(ctx, projectID, membershipID)
		if err != nil {
			return err
		}

		if err := s.membershipsRepo.Delete(ctx, projectID, membershipID); err != nil {
			return err
		}

		actorID := int(appctx.UserID(ctx))
		exec := db.TxFromContext(ctx)
		err = membershipaudit.Write(ctx, exec, string(old.ID), actorID, "delete", old, nil)
		if err != nil {
			return fmt.Errorf("write membership audit: %w", err)
		}

		//content := domain.UserNotificationContent{
		//	UserRemovedFromProject: &domain.UserRemovedFromProjectContent{
		//		ProjectName: project.Name,
		//		ByUser:      appctx.Username(ctx),
		//	},
		//}
		//err = s.userNotificationsUC.CreateNotification(
		//	ctx,
		//	membership.UserID,
		//	domain.UserNotificationTypeProjectRemoved,
		//	content,
		//)
		//if err != nil {
		//	return fmt.Errorf("create user notification: %w", err)
		//}

		return nil
	})
}
