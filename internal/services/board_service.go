package services

import (
	"context"
	"database/sql"
	"kanban-board/internal/repository"
	"kanban-board/internal/types"
	"kanban-board/internal/utils"
)

type BoardService interface {
	CreateBoard(ctx context.Context, b *types.Board) error
	GetBoardByID(ctx context.Context, id int) (*types.Board, error)
	GetBoards(ctx context.Context) ([]*types.Board, error)
	UpdateBoardName(ctx context.Context, id int, name string) error
	DeleteBoard(ctx context.Context, id int) error
}

type boardService struct {
	repo repository.BoardRepo
}

func NewBoardService(r repository.BoardRepo) BoardService {
	return &boardService{repo: r}
}

func (s *boardService) CreateBoard(ctx context.Context, b *types.Board) error {
	exists, err := s.repo.ProjectExists(ctx, b.ProjectID)
	if err != nil {
		return err
	}
	if !exists {
		return utils.ErrProjectNotFound
	}

	nameExists, err := s.repo.BoardNameExists(ctx, b.Name)
	if err != nil {
		return err
	}
	if nameExists {
		return utils.ErrBoardExists
	}

	return s.repo.Create(ctx, b)
}

func (s *boardService) GetBoardByID(ctx context.Context, id int) (*types.Board, error) {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrBoardNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *boardService) GetBoards(ctx context.Context) ([]*types.Board, error) {
	return s.repo.GetAll(ctx)
}

func (s *boardService) UpdateBoardName(ctx context.Context, id int, name string) error {
	if err := s.repo.Update(ctx, id, name); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrBoardNotFound
		}
		return err
	}
	return nil
}

func (s *boardService) DeleteBoard(ctx context.Context, id int) error {
	if id == 0 {
		return utils.ErrInvalidInput
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrBoardNotFound
		}
		return err
	}
	return nil
}
