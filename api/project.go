package api

import (
	"encoding/json"
	"kanban-board/models"
	"net/http"
)

func ListProjects(w http.ResponseWriter, r *http.Request) {
	response := models.ProjectResponse{Message: "Hello from list projects"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	response := models.ProjectResponse{Message: "Hello from create project"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}