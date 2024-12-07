package api

import (
	"database/sql"
	"encoding/json"
	"kanban-board/types"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *App) CreateColumnHandler(w http.ResponseWriter, r *http.Request) {
	var column types.Column

	err := json.NewDecoder(r.Body).Decode(&column)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if column.BoardID == 0 {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	var boardExists bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM boards WHERE id = $1)", column.BoardID).Scan(&boardExists)
	if err != nil {
		http.Error(w, "Database error while checking board", http.StatusInternalServerError)
		return
	}
	if !boardExists {
		http.Error(w, "Board does not exist", http.StatusNotFound)
		return
	}

	var existingColumnStatus bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM columns WHERE status = $1)", column.Status).Scan(&existingColumnStatus)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingColumnStatus {
		http.Error(w, "This column already exists", http.StatusConflict)
		return
	}

	err = app.DB.QueryRow("INSERT INTO columns (board_id, status) VALUES ($1, $2) RETURNING id",
		column.BoardID, column.Status).Scan(&column.ID)
	if err != nil {
		http.Error(w, "Error creating column", http.StatusInternalServerError)
		return
	}

	response := types.Column{
		ID:      column.ID,
		BoardID: column.BoardID,
		Status:  column.Status,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) GetColumnByID(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "id")
	var column types.Column

	err := app.DB.QueryRow("SELECT id, board_id, status FROM columns WHERE id = $1", columnID).
		Scan(&column.ID, &column.BoardID, &column.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Column not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(column)
}

func (app *App) GetColumns(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, board_id, status FROM columns")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var columns []types.Column
	for rows.Next() {
		var column types.Column
		if err := rows.Scan(&column.ID, &column.BoardID, &column.Status); err != nil {
			http.Error(w, "Error scanning projects", http.StatusInternalServerError)
			return
		}
		columns = append(columns, column)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(columns)
}

func (app *App) DeleteColumnHandler(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "id")
	if columnID == "" {
		http.Error(w, "Column ID is required", http.StatusBadRequest)
		return
	}

	result, err := app.DB.Exec("DELETE FROM columns WHERE id = $1", columnID)
	if err != nil {
		http.Error(w, "Database error while deleting column", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Column not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Column deleted successfully"))
}

func (app *App) UpdateColumnHandler(w http.ResponseWriter, r *http.Request) {
	columnID := chi.URLParam(r, "id")
	if columnID == "" {
		http.Error(w, "Column ID is required", http.StatusBadRequest)
		return
	}

	var updateData types.Column
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingColumn types.Column
	err = app.DB.QueryRow("SELECT id, board_id, status FROM columns WHERE id = $1",
		columnID).Scan(&existingColumn.ID, &existingColumn.BoardID, &existingColumn.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error while fetching column", http.StatusInternalServerError)
		return
	}

	_, err = app.DB.Exec("UPDATE columns SET status = $1 WHERE id = $2", updateData.Status, columnID)
	if err != nil {
		http.Error(w, "Error updating board name", http.StatusInternalServerError)
		return
	}

	response := types.Column{
		ID:      existingColumn.ID,
		BoardID: existingColumn.BoardID,
		Status:  updateData.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
