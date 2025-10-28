package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kanban-board/internal/services"
	"kanban-board/internal/types"
	"kanban-board/internal/utils"

	"github.com/go-chi/chi/v5"
)

type ProjectHandler struct {
	svc services.ProjectService
}

func NewProjectHandler(s services.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: s}
}

func (h *ProjectHandler) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var p types.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateProject(r.Context(), &p); err != nil {
		switch err {
		case utils.ErrInvalidInput:
			http.Error(w, "Invalid input", http.StatusBadRequest)
		case utils.ErrUserNotFound:
			http.Error(w, "User does not exists", http.StatusNotFound)
		case utils.ErrProjectExists:
			http.Error(w, "This project already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) GetProjectByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project id", http.StatusBadRequest)
		return
	}
	p, err := h.svc.GetProjectByID(r.Context(), id)
	if err != nil {
		if err == utils.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProjectHandler) GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetProjects(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *ProjectHandler) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project id", http.StatusBadRequest)
		return
	}
	var payload types.Project
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	payload.ID = id
	if err := h.svc.UpdateProject(r.Context(), &payload); err != nil {
		if err == utils.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func (h *ProjectHandler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteProject(r.Context(), id); err != nil {
		if err == utils.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
