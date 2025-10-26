package repository

import (
	"context"
	"database/sql"
	"kanban-board/internal/types"
	"time"
)

type TaskRepo interface {
	ColumnExists(ctx context.Context, columnID int) (bool, error)
	TaskTitleExists(ctx context.Context, title string) (bool, error)
	Create(ctx context.Context, t *types.Task) error
	GetByID(ctx context.Context, id int) (*types.Task, error)
	GetByColumn(ctx context.Context, columnID int) ([]*types.Task, error)
	GetAll(ctx context.Context) ([]*types.Task, error)
	GetTaskLogs(ctx context.Context, id int) ([]*types.TaskLog, error)
	Update(ctx context.Context, t *types.Task) error
	Delete(ctx context.Context, id int) error
}

type postgresTaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) TaskRepo {
	return &postgresTaskRepo{db: db}
}

func (r *postgresTaskRepo) ColumnExists(ctx context.Context, columnID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM columns WHERE id = $1)", columnID).Scan(&exists)
	return exists, err
}

func (r *postgresTaskRepo) TaskTitleExists(ctx context.Context, title string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE title = $1)", title).Scan(&exists)
	return exists, err
}

func (r *postgresTaskRepo) Create(ctx context.Context, t *types.Task) error {
	q := `INSERT INTO tasks (column_id, title, description) VALUES ($1, $2, $3) RETURNING id, created_at`
	if err := r.db.QueryRowContext(ctx, q, t.ColumnID, t.Title, t.Description).Scan(&t.ID, &t.CreatedAt); err != nil {
		return err
	}
	actionType := "create"
	logMessage := "Task created successfully"

	_, err := r.db.ExecContext(ctx, "INSERT INTO task_logs (task_id, action_type, log_message, created_at) VALUES ($1, $2, $3, $4)",
		t.ID, actionType, logMessage, t.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresTaskRepo) GetByID(ctx context.Context, id int) (*types.Task, error) {
	var t types.Task
	q := `SELECT id, column_id, title, description, created_at FROM tasks WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.ColumnID, &t.Title, &t.Description, &t.CreatedAt); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *postgresTaskRepo) GetByColumn(ctx context.Context, columnID int) ([]*types.Task, error) {
	q := `SELECT id, column_id, title, description, created_at FROM tasks WHERE column_id = $1`
	rows, err := r.db.QueryContext(ctx, q, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*types.Task
	for rows.Next() {
		var t types.Task
		if err := rows.Scan(&t.ID, &t.ColumnID, &t.Title, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}


func (r *postgresTaskRepo) GetAll(ctx context.Context) ([]*types.Task, error) {
	q := `SELECT id, column_id, title, description, created_at FROM tasks`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*types.Task
	for rows.Next() {
		var t types.Task
		if err := rows.Scan(&t.ID, &t.ColumnID, &t.Title, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *postgresTaskRepo) GetTaskLogs(ctx context.Context, id int) ([]*types.TaskLog, error) {
	q := `SELECT id, task_id, action_type, log_message, created_at FROM task_logs WHERE task_id = $1`
	rows, err := r.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*types.TaskLog
	for rows.Next() {
		var l types.TaskLog
		if err := rows.Scan(&l.ID, &l.TaskID, &l.ActionType, &l.LogMessage, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *postgresTaskRepo) Update(ctx context.Context, t *types.Task) error {
	if t.Title != "" {
		q := `UPDATE tasks SET title = $1 WHERE id = $2`
		_, err := r.db.ExecContext(ctx, q, t.Title, t.ID)
		if err != nil {
			return err
		}
	}

	if t.Description != "" {
		q := `UPDATE tasks SET description = $1 WHERE id = $2`
		_, err := r.db.ExecContext(ctx, q, t.Description, t.ID)
		if err != nil {
			return err
		}
	}

	actionType := "update"
	logMessage := "Task updated successfully"
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO task_logs (task_id, action_type, log_message, created_at) VALUES ($1, $2, $3, $4)`,
		t.ID, actionType, logMessage, time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresTaskRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
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

	actionType := "delete"
	logMessage := "Task deleted successfully"
	_, err = r.db.ExecContext(ctx, 
		"INSERT INTO task_logs (task_id, action_type, log_message, created_at) VALUES ($1, $2, $3, $4)",
		id, actionType, logMessage, time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}
