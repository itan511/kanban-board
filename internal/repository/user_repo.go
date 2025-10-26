package repository

import (
	"context"
	"database/sql"
	"kanban-board/internal/types"
)

type UserRepo interface {
	CreateUser(ctx context.Context, u *types.User) error
	GetByEmail(ctx context.Context, email string) (*types.User, error)
	GetByID(ctx context.Context, id int) (*types.User, error)
	UserExists(ctx context.Context, id int) (bool, error)

	ProjectExists(ctx context.Context, projectID int) (bool, error)
	UserInProject(ctx context.Context, projectID, userID int) (bool, error)
	AddUserToProject(ctx context.Context, projectID, userID int, role string) error
	RemoveUserFromProject(ctx context.Context, projectID, userID int) error
	GetProjectUsers(ctx context.Context, projectID int) ([]*types.ProjectUser, error)
}

type postgresUserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepo {
	return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) CreateUser(ctx context.Context, u *types.User) error {
	q := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`
	if err := r.db.QueryRowContext(ctx, q, u.Username, u.Email, u.Password).Scan(&u.ID, &u.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	var u types.User
	q := `SELECT id, username, email, password, created_at FROM users WHERE email = $1`
	if err := r.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id int) (*types.User, error) {
	var u types.User
	q := `SELECT id, username, email, password, created_at FROM users WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepo) UserExists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepo) ProjectExists(ctx context.Context, projectID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepo) UserInProject(ctx context.Context, projectID, userID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM project_users WHERE project_id = $1 AND user_id = $2)",
		projectID, userID).Scan(&exists)
	return exists, err
}

func (r *postgresUserRepo) AddUserToProject(ctx context.Context, projectID, userID int, role string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, $3)",
		projectID, userID, role)
	return err
}

func (r *postgresUserRepo) RemoveUserFromProject(ctx context.Context, projectID, userID int) error {
	res, err := r.db.ExecContext(ctx,
		"DELETE FROM project_users WHERE project_id = $1 AND user_id = $2", projectID, userID)
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

func (r *postgresUserRepo) GetProjectUsers(ctx context.Context, projectID int) ([]*types.ProjectUser, error) {
	q := `
		SELECT users.id, users.username, users.email, project_users.role
		FROM project_users
		JOIN users ON project_users.user_id = users.id
		WHERE project_users.project_id = $1
		ORDER BY users.username
	`
	rows, err := r.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*types.ProjectUser
	for rows.Next() {
		var pu types.ProjectUser
		if err := rows.Scan(&pu.ID, &pu.Username, &pu.Email, &pu.Role); err != nil {
			return nil, err
		}
		res = append(res, &pu)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
