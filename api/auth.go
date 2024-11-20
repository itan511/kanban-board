package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"kanban-board/types"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds types.Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingEmail string
	err = app.DB.QueryRow("SELECT email FROM users WHERE email = $1", creds.Email).Scan(&existingEmail)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingEmail != "" {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = app.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		creds.Username, creds.Email, string(hashedPassword),
	)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	tokenString, err := app.GenerateToken(creds.Email)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := types.UserResponse{
		Username: creds.Username,
		Email:    creds.Email,
		Token:    tokenString,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) GenerateToken(email string) (string, error) {
	JWTKey := []byte(os.Getenv("JWT_SECRET"))
	if len(JWTKey) == 0 {
		log.Fatalf("JWT_SECRET environment variable is not set")
	}

	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &types.Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(JWTKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	var creds types.Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var storedCreds types.Credentials

	err = app.DB.QueryRow("SELECT email, password FROM users WHERE email=$1",
		creds.Email).Scan(&storedCreds.Email, &storedCreds.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := app.GenerateToken(creds.Email)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := types.UserResponse{
		Username: creds.Username,
		Email:    creds.Email,
		Token:    tokenString,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}
