package repository

import (
	"context"
	"database/sql"
	"kanban-board/internal/types"
)

type ColumnRepo interface {
	BoardExists(ctx context.Context, boardID int) (bool, error)
	ColumnStatusExists(ctx context.Context, status string) (bool, error)
	Create(ctx context.Context, b *types.Column) error
	GetByID(ctx context.Context, id int) (*types.Column, error)
	GetAll(ctx context.Context) ([]*types.Column, error)
	Update(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
}

type postgresColumnRepo struct {
	db *sql.DB
}

func NewColumnRepo(db *sql.DB) ColumnRepo {
	return &postgresColumnRepo{db: db}
}

func (r *postgresColumnRepo) BoardExists(ctx context.Context, boardID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM boards WHERE id=$1)", boardID).Scan(&exists)
	return exists, err
}

func (r *postgresColumnRepo) ColumnStatusExists(ctx context.Context, status string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM columns WHERE status=$1)", status).Scan(&exists)
	return exists, err
}

func (r *postgresColumnRepo) Create(ctx context.Context, c *types.Column) error {
	q := `INSERT INTO columns (board_id, status) VALUES ($1, $2) RETURNING id`
	if err := r.db.QueryRowContext(ctx, q, c.BoardID, c.Status).Scan(&c.ID); err != nil {
		return err
	}
	return nil
}

func (r *postgresColumnRepo) GetByID(ctx context.Context, id int) (*types.Column, error) {
	var c types.Column
	q := `SELECT id, board_id, status FROM columns WHERE id=$1`
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&c.ID, &c.BoardID, &c.Status); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *postgresColumnRepo) GetAll(ctx context.Context) ([]*types.Column, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, board_id, status FROM columns")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*types.Column
	for rows.Next() {
		var c types.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Status); err != nil {
			return nil, err
		}
		columns = append(columns, &c)
	}
	return columns, rows.Err()
}

func (r *postgresColumnRepo) Update(ctx context.Context, id int, status string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE columns SET status=$1 WHERE id=$2", status, id)
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

func (r *postgresColumnRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM columns WHERE id=$1", id)
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
