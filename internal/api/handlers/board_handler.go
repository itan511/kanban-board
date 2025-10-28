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

type BoardHandler struct {
	svc services.BoardService
}

func NewBoardHandler(s services.BoardService) *BoardHandler {
	return &BoardHandler{svc: s}
}

func (h *BoardHandler) CreateBoardHandler(w http.ResponseWriter, r *http.Request) {
	var b types.Board
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := h.svc.CreateBoard(r.Context(), &b); err != nil {
		switch err {
		case utils.ErrProjectNotFound:
			http.Error(w, "Project does not exists", http.StatusNotFound)
		case utils.ErrBoardExists:
			http.Error(w, "This board already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func (h *BoardHandler) GetBoardByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid board id", http.StatusBadRequest)
		return
	}
	b, err := h.svc.GetBoardByID(r.Context(), id)
	if err != nil {
		if err == utils.ErrBoardNotFound {
			http.Error(w, "Board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func (h *BoardHandler) GetBoardsHandler(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetBoards(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *BoardHandler) UpdateBoardHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid board id", http.StatusBadRequest)
		return
	}
	var payload types.Board
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	payload.ID = id
	if err := h.svc.UpdateBoardName(r.Context(), id, payload.Name); err != nil {
		if err == utils.ErrBoardNotFound {
			http.Error(w, "Board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func (h *BoardHandler) DeleteBoardHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid board id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteBoard(r.Context(), id); err != nil {
		if err == utils.ErrBoardNotFound {
			http.Error(w, "`Board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
