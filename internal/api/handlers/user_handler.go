package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kanban-board/internal/services"
	"kanban-board/internal/utils"
	"kanban-board/internal/types"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	Svc services.UserService
}

func NewUserHandler(s services.UserService) *UserHandler {
	return &UserHandler{Svc: s}
}

func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds types.UserCreds
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.Svc.Register(r.Context(), &creds)
	if err != nil {
		switch err {
		case utils.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case utils.ErrUserExists:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds types.UserCreds
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.Svc.Login(r.Context(), &creds)
	if err != nil {
		switch err {
		case utils.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case utils.ErrInvalidCredentials:
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) AddUserToProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID == 0 {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var payload struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.UserID == 0 || payload.Role == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.Svc.AddUserToProject(r.Context(), projectID, payload.UserID, payload.Role); err != nil {
		switch err {
		case utils.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case utils.ErrProjectNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case utils.ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case utils.ErrUserExists:
			http.Error(w, "user is already in the project", http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user added to project"})
}

func (h *UserHandler) RemoveUserFromProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID == 0 {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var payload struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.UserID == 0 {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.Svc.RemoveUserFromProject(r.Context(), projectID, payload.UserID); err != nil {
		switch err {
		case utils.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case utils.ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "user removed from project"})
}

func (h *UserHandler) GetProjectUsersHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID == 0 {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	users, err := h.Svc.GetProjectUsers(r.Context(), projectID)
	if err != nil {
		switch err {
		case utils.ErrProjectNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case utils.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
