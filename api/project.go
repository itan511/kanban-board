package api

import (
	"database/sql"
	"encoding/json"
	"kanban-board/types"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *App) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var project types.Project

	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if project.UserID == 0 {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	var userExists bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", project.UserID).Scan(&userExists)
	if err != nil {
		http.Error(w, "Database error while checking user", http.StatusInternalServerError)
		return
	}
	if !userExists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	var existingProjectName string
	err = app.DB.QueryRow("SELECT name FROM projects WHERE name = $1", project.Name).Scan(&existingProjectName)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingProjectName != "" {
		http.Error(w, "This project already exists", http.StatusConflict)
		return
	}

	err = app.DB.QueryRow("INSERT INTO projects (name, user_id, description) VALUES ($1, $2, $3) RETURNING id, created_at",
		project.Name, project.UserID, project.Description).Scan(&project.ID, &project.CreatedAt)
	if err != nil {
		http.Error(w, "Error creating project", http.StatusInternalServerError)
		return
	}

	response := types.Project{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		UserID:      project.UserID,
		CreatedAt:   project.CreatedAt,
		Boards:      nil,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	var project types.Project

	err := app.DB.QueryRow("SELECT id, name, description, user_id, created_at FROM projects WHERE id = $1", projectID).
		Scan(&project.ID, &project.Name, &project.Description, &project.UserID, &project.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (app *App) GetProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, name, description, user_id, created_at FROM projects")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []types.Project
	for rows.Next() {
		var project types.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.UserID, &project.CreatedAt); err != nil {
			http.Error(w, "Error scanning projects", http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (app *App) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID проекта из параметров URL
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	// Выполняем запрос на удаление проекта
	result, err := app.DB.Exec("DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		http.Error(w, "Database error while deleting project", http.StatusInternalServerError)
		return
	}

	// Проверяем, был ли удален проект
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Project deleted successfully"))
}
