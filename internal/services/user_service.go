package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"kanban-board/internal/repository"
	"kanban-board/internal/utils"
	"kanban-board/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, creds *types.UserCreds) (*types.UserResponse, error)
	Login(ctx context.Context, creds *types.UserCreds) (*types.UserResponse, error)

	AddUserToProject(ctx context.Context, projectID, userID int, role string) error
	RemoveUserFromProject(ctx context.Context, projectID, userID int) error
	GetProjectUsers(ctx context.Context, projectID int) ([]*types.ProjectUser, error)
}

type userService struct {
	repo   repository.UserRepo
	jwtKey []byte
	tokenTTL time.Duration
}

func NewUserService(r repository.UserRepo, jwtKey []byte) UserService {
	return &userService{
		repo:     r,
		jwtKey:   jwtKey,
		tokenTTL: time.Hour,
	}
}

func toPublic(u *types.User) types.PublicUser {
	return types.PublicUser{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func (s *userService) generateToken(email string) (string, error) {
	if len(s.jwtKey) == 0 {
		return "", errors.New("jwt secret not configured")
	}
	exp := time.Now().Add(s.tokenTTL)
	claims := &types.Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

func (s *userService) Register(ctx context.Context, creds *types.UserCreds) (*types.UserResponse, error) {
	if creds == nil {
		return nil, utils.ErrInvalidInput
	}
	creds.Username = strings.TrimSpace(creds.Username)
	creds.Email = strings.TrimSpace(creds.Email)
	if creds.Email == "" || creds.Password == "" {
		return nil, utils.ErrInvalidInput
	}

	_, err := s.repo.GetByEmail(ctx, creds.Email)
	if err == nil {
		return nil, utils.ErrUserExists
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &types.User{
		Username: creds.Username,
		Email:    creds.Email,
		Password: string(hashed),
	}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, err
	}

	token, err := s.generateToken(u.Email)
	if err != nil {
		return nil, err
	}

	return &types.UserResponse{
		User:  toPublic(u),
		Token: token,
	}, nil
}

func (s *userService) Login(ctx context.Context, creds *types.UserCreds) (*types.UserResponse, error) {
	if creds == nil || creds.Email == "" || creds.Password == "" {
		return nil, utils.ErrInvalidInput
	}

	u, err := s.repo.GetByEmail(ctx, creds.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(creds.Password)); err != nil {
		return nil, utils.ErrInvalidCredentials
	}

	token, err := s.generateToken(u.Email)
	if err != nil {
		return nil, err
	}

	return &types.UserResponse{
		User:  toPublic(u),
		Token: token,
	}, nil
}

func (s *userService) AddUserToProject(ctx context.Context, projectID, userID int, role string) error {
	if projectID == 0 || userID == 0 || strings.TrimSpace(role) == "" {
		return utils.ErrInvalidInput
	}

	ok, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return err
	}
	if !ok {
		return utils.ErrProjectNotFound
	}

	ok, err = s.repo.UserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return utils.ErrUserNotFound
	}

	in, err := s.repo.UserInProject(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if in {
		return utils.ErrUserExists 	
	}

	if err := s.repo.AddUserToProject(ctx, projectID, userID, role); err != nil {
		return err
	}
	return nil
}

func (s *userService) RemoveUserFromProject(ctx context.Context, projectID, userID int) error {
	if projectID == 0 || userID == 0 {
		return utils.ErrInvalidInput
	}

	in, err := s.repo.UserInProject(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !in {
		return utils.ErrUserNotFound
	}

	if err := s.repo.RemoveUserFromProject(ctx, projectID, userID); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (s *userService) GetProjectUsers(ctx context.Context, projectID int) ([]*types.ProjectUser, error) {
	if projectID == 0 {
		return nil, utils.ErrInvalidInput
	}
	ok, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, utils.ErrProjectNotFound
	}

	users, err := s.repo.GetProjectUsers(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return users, nil
}
