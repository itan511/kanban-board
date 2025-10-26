package repository

import (
	"context"
	"database/sql"
	"kanban-board/internal/types"
)

type BoardRepo interface {
	ProjectExists(ctx context.Context, projectID int) (bool, error)
	BoardNameExists(ctx context.Context, name string) (bool, error)
	Create(ctx context.Context, b *types.Board) error
	GetByID(ctx context.Context, id int) (*types.Board, error)
	GetAll(ctx context.Context) ([]*types.Board, error)
	Update(ctx context.Context, id int, name string) error
	Delete(ctx context.Context, id int) error
}

type postgresBoardRepo struct {
	db *sql.DB
}

func NewBoardRepo(db *sql.DB) BoardRepo {
	return &postgresBoardRepo{db: db}
}

func (r *postgresBoardRepo) ProjectExists(ctx context.Context, projectID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)", projectID).Scan(&exists)
	return exists, err
}

func (r *postgresBoardRepo) BoardNameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM boards WHERE name=$1)", name).Scan(&exists)
	return exists, err
}

func (r *postgresBoardRepo) Create(ctx context.Context, b *types.Board) error {
	q := `INSERT INTO boards (project_id, name) VALUES ($1, $2) RETURNING id`
	if err := r.db.QueryRowContext(ctx, q, b.ProjectID, b.Name).Scan(&b.ID); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, `INSERT INTO columns (board_id, status) VALUES
		($1, 'todo'),
		($1, 'doing'),
		($1, 'done')`, b.ID)
	return err
}

func (r *postgresBoardRepo) GetByID(ctx context.Context, id int) (*types.Board, error) {
	var b types.Board
	q := `SELECT id, project_id, name FROM boards WHERE id=$1`
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&b.ID, &b.ProjectID, &b.Name); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *postgresBoardRepo) GetAll(ctx context.Context) ([]*types.Board, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, project_id, name FROM boards")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []*types.Board
	for rows.Next() {
		var b types.Board
		if err := rows.Scan(&b.ID, &b.ProjectID, &b.Name); err != nil {
			return nil, err
		}
		boards = append(boards, &b)
	}
	return boards, rows.Err()
}

func (r *postgresBoardRepo) Update(ctx context.Context, id int, name string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE boards SET name=$1 WHERE id=$2", name, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *postgresBoardRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM boards WHERE id=$1", id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}
