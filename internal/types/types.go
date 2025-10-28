package types

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ----------------------------
// Auth / User related types
// ----------------------------

// User — доменная модель пользователя, используется в сервисах и репозиториях.
// Поле Password хранится захэшированным и НЕ должно попадать в JSON-ответы.
type User struct {
	ID        int       `json:"id"`                  // DB id
	Username  string    `json:"username,omitempty"`  // отображаемое имя
	Email     string    `json:"email"`               // уникальный email
	Password  string    `json:"-"`                   // хеш пароля, не сериализуется в json
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// UserCreds — структура, используемая для регистрации и логина (входные данные).
// Для регистрации может использоваться поле Username (опционально).
type UserCreds struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate проверяет минимальные требования к UserCreds.
// Удобно вызывать в сервисе: if err := creds.Validate(); err != nil { return err }
func (c *UserCreds) Validate() error {
	if c == nil {
		return errors.New("credentials are nil")
	}
	if strings.TrimSpace(c.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(c.Password) == "" {
		return errors.New("password is required")
	}
	return nil
}

// PublicUser — DTO, которое безопасно отдавать клиенту (без поля Password).
// Хендлеры и ответы API должны использовать именно этот тип для вывода информации о пользователе.
type PublicUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// UserResponse — отдаётся клиенту после регистрации/логина (включает JWT).
type UserResponse struct {
	User  PublicUser `json:"user"`
	Token string     `json:"token"`
}

// ProjectUser — информация о пользователе в контексте проекта (role)
type ProjectUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// ----------------------------
// JWT Claims
// ----------------------------

// Claims — кастомные claims для JWT (заполнять в UserService при генерации токена)
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// ----------------------------
// Остальные типы (без изменений, для контекста)
// ----------------------------

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
