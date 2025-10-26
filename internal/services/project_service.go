package services

import (
	"context"
	"database/sql"
	"kanban-board/internal/repository"
	"kanban-board/internal/types"
	"kanban-board/internal/utils"
)

type ProjectService interface {
	CreateProject(ctx context.Context, p *types.Project) error
	GetProjectByID(ctx context.Context, id int) (*types.Project, error)
	GetProjects(ctx context.Context) ([]*types.Project, error)
	UpdateProject(ctx context.Context, p *types.Project) error
	DeleteProject(ctx context.Context, id int) error
}

type projectService struct {
	repo repository.ProjectRepo
}

func NewProjectService(r repository.ProjectRepo) ProjectService {
	return &projectService{repo: r}
}

func (s *projectService) CreateProject(ctx context.Context, p *types.Project) error {
	if p == nil || p.Name == "" {
		return utils.ErrInvalidInput
	}

	ok, err := s.repo.UserExists(ctx, p.UserID)
	if err != nil {
		return err
	}
	if !ok {
		return utils.ErrUserNotFound
	}

	exists, err := s.repo.ProjectNameExists(ctx, p.Name)
	if err != nil {
		return err
	}
	if exists {
		return utils.ErrProjectExists
	}

	return s.repo.Create(ctx, p)
}

func (s *projectService) GetProjectByID(ctx context.Context, id int) (*types.Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrProjectNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *projectService) GetProjects(ctx context.Context) ([]*types.Project, error) {
	return s.repo.GetAll(ctx)
}

func (s *projectService) UpdateProject(ctx context.Context, p *types.Project) error {
	if p == nil || p.ID == 0 || p.Name == "" {
		return utils.ErrInvalidInput
	}
	if err := s.repo.Update(ctx, p); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrProjectNotFound
		}
		return err
	}
	return nil
}

func (s *projectService) DeleteProject(ctx context.Context, id int) error {
	if id == 0 {
		return utils.ErrInvalidInput
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrProjectNotFound
		}
		return err
	}
	return nil
}
