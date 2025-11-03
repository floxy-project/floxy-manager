package projects

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/db"
)

type ProjectService struct {
	txManager   db.TxManager
	projectRepo contract.ProjectsRepository
}

func New(
	txManager db.TxManager,
	projectRepo contract.ProjectsRepository,
) *ProjectService {
	return &ProjectService{
		txManager:   txManager,
		projectRepo: projectRepo,
	}
}

func (s *ProjectService) GetProject(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

func (s *ProjectService) CreateProject(
	ctx context.Context,
	name, description string,
) (domain.Project, error) {
	project := domain.ProjectDTO{
		Name:        name,
		Description: description,
	}

	var id domain.ProjectID
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var err error
		id, err = s.projectRepo.Create(ctx, &project)
		if err != nil {
			return err
		}

		// Create tags from system categories
		// err = s.tagsUseCase.CreateTagsFromCategories(ctx, id)
		// if err != nil {
		//	slog.Error("failed to create tags from categories", "error", err, "project_id", id)
		//	// Don't fail project creation if tag creation fails
		//}

		return nil
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("create project: %w", err)
	}

	return domain.Project{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *ProjectService) List(ctx context.Context) ([]domain.Project, error) {
	return s.projectRepo.List(ctx)
}

func (s *ProjectService) UpdateInfo(
	ctx context.Context,
	id domain.ProjectID,
	name, description string,
) (domain.Project, error) {
	// Check if the project exists
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Project{}, fmt.Errorf("failed to get project: %w", err)
	}

	// Update the project
	err = s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		return s.projectRepo.Update(ctx, id, name, description)
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("failed to update project: %w", err)
	}

	// Return the updated project with extended info
	project.Name = name
	project.Description = description

	return project, nil
}

func (s *ProjectService) ArchiveProject(ctx context.Context, id domain.ProjectID) error {
	// Check if the project exists
	_, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Archive the project
	err = s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		return s.projectRepo.Archive(ctx, id)
	})
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	slog.Info("project archived", "project_id", id)

	return nil
}
