package services

import (
	"context"
	"database/sql"
	"kanban-board/internal/repository"
	"kanban-board/internal/types"
	"kanban-board/internal/utils"
	"strings"
)

type TaskService interface {
	CreateTask(ctx context.Context, t *types.Task) error
	GetTaskByID(ctx context.Context, id int) (*types.Task, error)
	GetTasks(ctx context.Context) ([]*types.Task, error)
	GetTasksByColumn(ctx context.Context, columnID int) ([]*types.Task, error)
	GetTaskLogs(ctx context.Context, id int) ([]*types.TaskLog, error)
	UpdateTask(ctx context.Context, t *types.Task) error
	DeleteTask(ctx context.Context, id int) error
}

type taskService struct {
	repo repository.TaskRepo
}

func NewTaskService(r repository.TaskRepo) TaskService {
	return &taskService{repo: r}
}

func (s *taskService) CreateTask(ctx context.Context, t *types.Task) error {
	if t == nil {
		return utils.ErrInvalidInput
	}
	if t.ColumnID == 0 || strings.TrimSpace(t.Title) == "" {
		return utils.ErrInvalidInput
	}
	exists, err := s.repo.ColumnExists(ctx, t.ColumnID)
	if err != nil {
		return err
	}
	if !exists {
		return utils.ErrColumnNotFound
	}

	titleExists, err := s.repo.TaskTitleExists(ctx, t.Title)
	if err != nil {
		return err
	}
	if titleExists {
		return utils.ErrTaskExists
	}

	return s.repo.Create(ctx, t)
}

func (s *taskService) GetTaskByID(ctx context.Context, id int) (*types.Task, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrTaskNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *taskService) GetTasks(ctx context.Context) ([]*types.Task, error) {
	return s.repo.GetAll(ctx)
}

func (s *taskService) GetTasksByColumn(ctx context.Context, columnID int) ([]*types.Task, error) {
	if columnID == 0 {
		return nil, utils.ErrInvalidInput
	}
	tasks, err := s.repo.GetByColumn(ctx, columnID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) GetTaskLogs(ctx context.Context, id int) ([]*types.TaskLog, error) {
	if id == 0 {
		return nil, utils.ErrInvalidInput
	}
	logs, err := s.repo.GetTaskLogs(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrTaskNotFound
		}
		return nil, err
	}
	return logs, nil
}

func (s *taskService) UpdateTask(ctx context.Context, t *types.Task) error {
	if t == nil || t.ID == 0 {
		return utils.ErrInvalidInput
	}
	if err := s.repo.Update(ctx, t); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrTaskNotFound
		}
		return err
	}
	return nil
}

func (s *taskService) DeleteTask(ctx context.Context, id int) error {
	if id == 0 {
		return utils.ErrInvalidInput
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrTaskNotFound
		}
		return err
	}
	return nil
}
