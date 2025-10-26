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

type ColumnHandler struct {
	svc services.ColumnService
}

func NewColumnHandler(s services.ColumnService) *ColumnHandler {
	return &ColumnHandler{svc: s}
}

func (h *ColumnHandler) CreateColumnHandler(w http.ResponseWriter, r *http.Request) {
	var c types.Column
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateColumn(r.Context(), &c); err != nil {
		switch err {
		case utils.ErrBoardNotFound:
			http.Error(w, "Board does not exists", http.StatusNotFound)
		case utils.ErrColumnExists:
			http.Error(w, "This column already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *ColumnHandler) GetColumnByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid column id", http.StatusBadRequest)
		return
	}
	c, err := h.svc.GetColumnByID(r.Context(), id)
	if err != nil {
		if err == utils.ErrColumnNotFound {
			http.Error(w, "Column not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *ColumnHandler) GetColumnsHandler(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetColumns(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *ColumnHandler) UpdateColumnHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid column id", http.StatusBadRequest)
		return
	}
	var payload types.Column
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	payload.ID = id
	if err := h.svc.UpdateColumnStatus(r.Context(), id, payload.Status); err != nil {
		if err == utils.ErrColumnNotFound {
			http.Error(w, "Column not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func (h *ColumnHandler) DeleteColumnHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid column id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteColumn(r.Context(), id); err != nil {
		if err == utils.ErrBoardNotFound {
			http.Error(w, "`Board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
