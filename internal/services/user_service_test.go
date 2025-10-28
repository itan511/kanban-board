package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"kanban-board/internal/types"
	"kanban-board/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getByEmailResult *types.User
	getByEmailErr    error
	createErr        error
}

func (m *mockUserRepo) CreateUser(ctx context.Context, u *types.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	u.ID = 123
	u.CreatedAt = time.Now()
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	if m.getByEmailResult == nil {
		return nil, sql.ErrNoRows
	}
	return m.getByEmailResult, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*types.User, error) { return nil, nil }
func (m *mockUserRepo) UserExists(ctx context.Context, id int) (bool, error) { return false, nil }
func (m *mockUserRepo) AddUserToProject(ctx context.Context, projectID int, userID int, role string) error { return nil }
func (m *mockUserRepo) ProjectExists(ctx context.Context, projectID int) (bool, error) { return false, nil }
func (m *mockUserRepo) UserInProject(ctx context.Context, projectID, userID int) (bool, error) { return false, nil }
func (m *mockUserRepo) RemoveUserFromProject(ctx context.Context, projectID, userID int) error { return nil }
func (m *mockUserRepo) GetProjectUsers(ctx context.Context, projectID int) ([]*types.ProjectUser, error) { return nil, nil }

func TestUserService_Register_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := &mockUserRepo{
		getByEmailResult: nil,
		createErr:        nil,
	}
	svc := &userService{repo: mockRepo, jwtKey: []byte("secret"), tokenTTL: time.Hour}

	creds := &types.UserCreds{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	}

	resp, err := svc.Register(ctx, creds)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil || resp.User.Email != "alice@example.com" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if resp.Token == "" {
		t.Fatalf("expected token, got empty")
	}
}

func TestUserService_Register_ExistingEmail(t *testing.T) {
	ctx := context.Background()

	mockRepo := &mockUserRepo{
		getByEmailResult: &types.User{ID: 1, Email: "a@a.com"},
	}
	svc := &userService{repo: mockRepo, jwtKey: []byte("secret"), tokenTTL: time.Hour}

	creds := &types.UserCreds{Email: "a@a.com", Password: "p"}
	_, err := svc.Register(ctx, creds)
	if !errors.Is(err, utils.ErrUserExists) {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestUserService_Login_Success(t *testing.T) {
	ctx := context.Background()
	hashed, _ := func() (string, error) {
		h, err := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		return string(h), err
	}()
	mockRepo := &mockUserRepo{
		getByEmailResult: &types.User{ID: 1, Email: "bob@example.com", Password: hashed},
	}
	svc := &userService{repo: mockRepo, jwtKey: []byte("secret"), tokenTTL: time.Hour}

	creds := &types.UserCreds{Email: "bob@example.com", Password: "pass"}
	resp, err := svc.Login(ctx, creds)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected token")
	}
}

func TestUserService_Login_InvalidCredentials(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockUserRepo{
		getByEmailErr: sql.ErrNoRows,
	}
	svc := &userService{repo: mockRepo, jwtKey: []byte("secret"), tokenTTL: time.Hour}
	_, err := svc.Login(ctx, &types.UserCreds{Email: "x", Password: "y"})
	if !errors.Is(err, utils.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
