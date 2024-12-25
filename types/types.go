package types

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserCreds struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type UserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type Project struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Boards      []Board   `json:"boards,omitempty"`
}

type Board struct {
	ID        int      `json:"id"`
	ProjectID int      `json:"project_id"`
	Name      string   `json:"name"`
	Columns   []Column `json:"columns,omitempty"`
}

type Column struct {
	ID      int    `json:"id"`
	BoardID int    `json:"board_id"`
	Status  string `json:"status"`
	Tasks   []Task `json:"tasks,omitempty"`
}

type Task struct {
	ID          int       `json:"id"`
	ColumnID    int       `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	TaskLogs    []TaskLog `json:"task_logs,omitempty"`
}

type TaskLog struct {
	ID         int       `json:"id"`
	TaskID     int       `json:"task_id"`
	ActionType string    `json:"action_type"`
	LogMessage string    `json:"log_message"`
	CreatedAt  time.Time `json:"created_at"`
}
