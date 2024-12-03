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

	_, err = app.DB.Exec("INSERT INTO project_users (project_id, user_id, role) VALUES ($1, $2, 'user')",
		project.ID, project.UserID)
	if err != nil {
		http.Error(w, "Error adding user to project", http.StatusInternalServerError)
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
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	result, err := app.DB.Exec("DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		http.Error(w, "Database error while deleting project", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Project deleted successfully"))
}

func (app *App) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	var updateData types.Project
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingProject types.Project
	err = app.DB.QueryRow("SELECT id, name, description, user_id, created_at FROM projects WHERE id = $1",
		projectID).Scan(&existingProject.ID, &existingProject.Name, &existingProject.Description, &existingProject.UserID, &existingProject.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error while fetching project", http.StatusInternalServerError)
		return
	}

	if updateData.Name != "" {
		_, err = app.DB.Exec("UPDATE projects SET name = $1 WHERE id = $2", updateData.Name, projectID)
		if err != nil {
			http.Error(w, "Error updating project name", http.StatusInternalServerError)
			return
		}
		existingProject.Name = updateData.Name
	}

	if updateData.Description != "" {
		_, err = app.DB.Exec("UPDATE projects SET description = $1 WHERE id = $2", updateData.Description, projectID)
		if err != nil {
			http.Error(w, "Error updating project description", http.StatusInternalServerError)
			return
		}
		existingProject.Description = updateData.Description
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingProject)
}
