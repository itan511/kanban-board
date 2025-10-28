package services

import (
	"context"
	"database/sql"
	"kanban-board/internal/repository"
	"kanban-board/internal/types"
	"kanban-board/internal/utils"
)

type ColumnService interface {
	CreateColumn(ctx context.Context, c *types.Column) error
	GetColumnByID(ctx context.Context, id int) (*types.Column, error)
	GetColumns(ctx context.Context) ([]*types.Column, error)
	UpdateColumnStatus(ctx context.Context, id int, status string) error
	DeleteColumn(ctx context.Context, id int) error
}

type columnService struct {
	repo repository.ColumnRepo
}

func NewColumnService(r repository.ColumnRepo) ColumnService {
	return &columnService{repo: r}
}

func (s *columnService) CreateColumn(ctx context.Context, c *types.Column) error {
	exists, err := s.repo.BoardExists(ctx, c.BoardID)
	if err != nil {
		return err
	}
	if !exists {
		return utils.ErrBoardNotFound
	}

	nameExists, err := s.repo.ColumnStatusExists(ctx, c.Status)
	if err != nil {
		return err
	}
	if nameExists {
		return utils.ErrColumnExists
	}

	return s.repo.Create(ctx, c)
}

func (s *columnService) GetColumnByID(ctx context.Context, id int) (*types.Column, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrColumnNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *columnService) GetColumns(ctx context.Context) ([]*types.Column, error) {
	return s.repo.GetAll(ctx)
}

func (s *columnService) UpdateColumnStatus(ctx context.Context, id int, status string) error {
	if err := s.repo.Update(ctx, id, status); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrColumnNotFound
		}
		return err
	}
	return nil
}

func (s *columnService) DeleteColumn(ctx context.Context, id int) error {
	if id == 0 {
		return utils.ErrInvalidInput
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrColumnNotFound
		}
		return err
	}
	return nil
}
