package repository

import (
	"context"
	"database/sql"
	"kanban-board/internal/types"
)

type ProjectRepo interface {
	UserExists(ctx context.Context, userID int) (bool, error)
	ProjectNameExists(ctx context.Context, name string) (bool, error)
	Create(ctx context.Context, p *types.Project) error
	GetByID(ctx context.Context, id int) (*types.Project, error)
	GetAll(ctx context.Context) ([]*types.Project, error)
	Update(ctx context.Context, p *types.Project) error
	Delete(ctx context.Context, id int) error
}

type postgresProjectRepo struct {
	db *sql.DB
}

func NewProjectRepo(db *sql.DB) ProjectRepo {
	return &postgresProjectRepo{db: db}
}

func (r *postgresProjectRepo) UserExists(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	return exists, err
}

func (r *postgresProjectRepo) ProjectNameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE name = $1)", name).Scan(&exists)
	return exists, err
}

func (r *postgresProjectRepo) Create(ctx context.Context, p *types.Project) error {
	q := `INSERT INTO projects (name, user_id, description) VALUES ($1, $2, $3) RETURNING id, created_at`
	if err := r.db.QueryRowContext(ctx, q, p.Name, p.UserID, p.Description).Scan(&p.ID, &p.CreatedAt); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, $3)",
		p.ID, p.UserID, "user",
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresProjectRepo) GetByID(ctx context.Context, id int) (*types.Project, error) {
	var p types.Project
	q := `SELECT id, name, description, user_id, created_at FROM projects WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.Name, &p.Description, &p.UserID, &p.CreatedAt); err != nil {
		return nil, err 
	}
	return &p, nil
}

func (r *postgresProjectRepo) GetAll(ctx context.Context) ([]*types.Project, error) {
	q := `SELECT id, name, description, user_id, created_at FROM projects ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*types.Project
	for rows.Next() {
		var p types.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.UserID, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *postgresProjectRepo) Update(ctx context.Context, p *types.Project) error {
	q := `UPDATE projects SET name = $1, description = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, q, p.Name, p.Description, p.ID)
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

func (r *postgresProjectRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
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

