package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *App) AddUserToProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	var userData struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"` // Роль пользователя в проекте
	}

	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil || userData.UserID == 0 || userData.Role == "" {
		http.Error(w, "Invalid request payload or missing fields", http.StatusBadRequest)
		return
	}

	// Проверка существования проекта
	var projectExists bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&projectExists)
	if err != nil {
		http.Error(w, "Database error while checking project", http.StatusInternalServerError)
		return
	}
	if !projectExists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Проверка существования пользователя
	var userExists bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userData.UserID).Scan(&userExists)
	if err != nil {
		http.Error(w, "Database error while checking user", http.StatusInternalServerError)
		return
	}
	if !userExists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Проверка, есть ли уже пользователь в проекте
	var userAlreadyInProject bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM project_users WHERE project_id = $1 AND user_id = $2)",
		projectID, userData.UserID).Scan(&userAlreadyInProject)
	if err != nil {
		http.Error(w, "Database error while checking user in project", http.StatusInternalServerError)
		return
	}
	if userAlreadyInProject {
		http.Error(w, "User is already in the project", http.StatusConflict)
		return
	}

	// Добавление пользователя в проект
	_, err = app.DB.Exec("INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, $3)",
		projectID, userData.UserID, userData.Role)
	if err != nil {
		http.Error(w, "Error adding user to project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User added to project successfully"))
}

func (app *App) RemoveUserFromProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	var userData struct {
		UserID int64 `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil || userData.UserID == 0 {
		http.Error(w, "Invalid request payload or missing user_id", http.StatusBadRequest)
		return
	}

	// Проверяем, находится ли пользователь в проекте
	var userExistsInProject bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM project_users WHERE project_id = $1 AND user_id = $2)",
		projectID, userData.UserID).Scan(&userExistsInProject)
	if err != nil {
		http.Error(w, "Database error while checking user in project", http.StatusInternalServerError)
		return
	}
	if !userExistsInProject {
		http.Error(w, "User is not in the project", http.StatusNotFound)
		return
	}

	// Удаляем пользователя из проекта
	_, err = app.DB.Exec("DELETE FROM project_users WHERE project_id = $1 AND user_id = $2", projectID, userData.UserID)
	if err != nil {
		http.Error(w, "Error removing user from project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User removed from project successfully"))
}

func (app *App) GetProjectUsersHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	rows, err := app.DB.Query(`
		SELECT id, username, email, role
		FROM project_users
		JOIN users ON project_users.user_id = users.id
		WHERE project_users.project_id = $1`, projectID)
	if err != nil {
		http.Error(w, "Database error while fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	for rows.Next() {
		var user struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Database error after fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
