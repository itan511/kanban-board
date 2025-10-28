package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"kanban-board/internal/types"
	"kanban-board/internal/utils"
)

type mockProjectRepo struct {
	userExistsRes   bool
	userExistsErr   error
	existsByNameRes bool
	existsByNameErr error
	createErr       error
	getByIDRes      *types.Project
	getByIDErr      error
}

func (m *mockProjectRepo) UserExists(ctx context.Context, userID int) (bool, error) {
	return m.userExistsRes, m.userExistsErr
}
func (m *mockProjectRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	return m.existsByNameRes, m.existsByNameErr
}
func (m *mockProjectRepo) Create(ctx context.Context, p *types.Project) error {
	if m.createErr != nil {
		return m.createErr
	}
	p.ID = 10
	p.CreatedAt = time.Now()
	return nil
}
func (m *mockProjectRepo) GetByID(ctx context.Context, id int) (*types.Project, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDRes, nil
}
func (m *mockProjectRepo) GetAll(ctx context.Context) ([]*types.Project, error) { return nil, nil }
func (m *mockProjectRepo) Update(ctx context.Context, p *types.Project) error   { return nil }
func (m *mockProjectRepo) Delete(ctx context.Context, id int) error             { return nil }
func (m *mockProjectRepo) ProjectNameExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func TestProjectService_Create_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockProjectRepo{userExistsRes: true, existsByNameRes: false}
	svc := &projectService{repo: mock}

	p := &types.Project{UserID: 1, Name: "New Project"}
	err := svc.CreateProject(ctx, p)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p.ID == 0 {
		t.Fatalf("expected id set")
	}
}

func TestProjectService_Create_UserNotExists(t *testing.T) {
	ctx := context.Background()
	mock := &mockProjectRepo{userExistsRes: false}
	svc := &projectService{repo: mock}
	err := svc.CreateProject(ctx, &types.Project{UserID: 999, Name: "x"})
	if !errors.Is(err, utils.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestProjectService_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mock := &mockProjectRepo{getByIDErr: sql.ErrNoRows}
	svc := &projectService{repo: mock}
	_, err := svc.GetProjectByID(ctx, 42)
	if !errors.Is(err, utils.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}
