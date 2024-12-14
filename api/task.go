package api

import (
	"database/sql"
	"encoding/json"
	"kanban-board/types"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (app *App) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task types.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if task.ColumnID == 0 {
		http.Error(w, "ColumnID is required", http.StatusBadRequest)
		return
	}

	var columnExists bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM columns WHERE id = $1)", task.ColumnID).Scan(&columnExists)
	if err != nil {
		http.Error(w, "Database error while checking column", http.StatusInternalServerError)
		return
	}
	if !columnExists {
		http.Error(w, "Column does not exist", http.StatusNotFound)
		return
	}

	var existingTitle string
	err = app.DB.QueryRow("SELECT title FROM tasks WHERE title = $1", task.Title).Scan(&existingTitle)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingTitle != "" {
		http.Error(w, "This task already exists", http.StatusConflict)
		return
	}

	err = app.DB.QueryRow("INSERT INTO tasks (column_id, title, description) VALUES ($1, $2, $3) RETURNING id, created_at",
		task.ColumnID, task.Title, task.Description).Scan(&task.ID, &task.CreatedAt)
	if err != nil {
		http.Error(w, "Error creating task", http.StatusInternalServerError)
		return
	}

	actionType := "create"
	logMessage := "Task created successfully"

	_, err = app.DB.Exec(
		"INSERT INTO task_logs (task_id, action_type, log_message, created_at) VALUES ($1, $2, $3, $4)",
		task.ID, actionType, logMessage, task.CreatedAt,
	)
	if err != nil {
		http.Error(w, "Error logging task creation", http.StatusInternalServerError)
		return
	}

	response := types.Task{
		ID:          task.ID,
		ColumnID:    task.ColumnID,
		Title:       task.Title,
		Description: task.Description,
		CreatedAt:   task.CreatedAt,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	var task types.Task

	err := app.DB.QueryRow("SELECT id, column_id, title, description, created_at FROM tasks WHERE id = $1", taskID).
		Scan(&task.ID, &task.ColumnID, &task.Title, &task.Description, &task.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (app *App) GetTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, column_id, title, description, created_at FROM tasks")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []types.Task
	for rows.Next() {
		var task types.Task
		if err := rows.Scan(&task.ID, &task.ColumnID, &task.Title, &task.Description, &task.CreatedAt); err != nil {
			http.Error(w, "Error scanning tasks", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (app *App) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	result, err := app.DB.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		http.Error(w, "Database error while deleting task", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	actionType := "delete"
	logMessage := "Task deleted successfully"
	_, err = app.DB.Exec("INSERT INTO task_logs (task_id, action_type, log_message, created_at) VALUES ($1, $2, $3, $4)",
		taskID, actionType, logMessage, time.Now(),
	)
	if err != nil {
		http.Error(w, "Error logging task deletion", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task deleted successfully"))
}

func (app *App) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	var updateData types.Task
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingTask types.Task
	err = app.DB.QueryRow("SELECT id, column_id, title, description, created_at FROM tasks WHERE id = $1",
		taskID).Scan(&existingTask.ID, &existingTask.ColumnID, &existingTask.Title, &existingTask.Description, &existingTask.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error while fetching project", http.StatusInternalServerError)
		return
	}

	if updateData.Title != "" {
		_, err = app.DB.Exec("UPDATE tasks SET title = $1 WHERE id = $2", updateData.Title, taskID)
		if err != nil {
			http.Error(w, "Error updating task title", http.StatusInternalServerError)
			return
		}
		existingTask.Title = updateData.Title
	}

	if updateData.Description != "" {
		_, err = app.DB.Exec("UPDATE tasks SET description = $1 WHERE id = $2", updateData.Description, taskID)
		if err != nil {
			http.Error(w, "Error updating task description", http.StatusInternalServerError)
			return
		}
		existingTask.Description = updateData.Description
	}

	actionType := "update"
	logMessage := "Task updated successfully"
	_, err = app.DB.Exec("INSERT INTO task_logs (task_id, action_type, log_message, created_at) VALUES ($1, $2, $3, $4)",
		taskID, actionType, logMessage, time.Now(),
	)
	if err != nil {
		http.Error(w, "Error logging task updation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTask)
}
