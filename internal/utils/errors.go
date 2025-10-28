package utils

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrProjectNotFound    = errors.New("project not found")
	ErrProjectExists      = errors.New("project already exists")
	ErrBoardNotFound      = errors.New("board not found")
	ErrBoardExists        = errors.New("board already exists")
	ErrColumnNotFound     = errors.New("column not found")
	ErrColumnExists       = errors.New("column already exists")
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskExists         = errors.New("task already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
