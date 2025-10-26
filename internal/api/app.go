package api

import (
	"database/sql"
)

type App struct {
	DB     *sql.DB
	JWTKey []byte
}
