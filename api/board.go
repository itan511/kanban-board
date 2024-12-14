package api

import (
	"database/sql"
	"encoding/json"
	"kanban-board/types"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *App) CreateBoardHandler(w http.ResponseWriter, r *http.Request) {
	var board types.Board

	err := json.NewDecoder(r.Body).Decode(&board)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if board.ProjectID == 0 {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	var projectExists bool
	err = app.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", board.ProjectID).Scan(&projectExists)
	if err != nil {
		http.Error(w, "Database error while checking project", http.StatusInternalServerError)
		return
	}
	if !projectExists {
		http.Error(w, "Project does not exist", http.StatusNotFound)
		return
	}

	var existingBoardName string
	err = app.DB.QueryRow("SELECT name FROM boards WHERE name = $1", board.Name).Scan(&existingBoardName)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingBoardName != "" {
		http.Error(w, "This board already exists", http.StatusConflict)
		return
	}

	err = app.DB.QueryRow("INSERT INTO boards (project_id, name) VALUES ($1, $2) RETURNING id",
		board.ProjectID, board.Name).Scan(&board.ID)
	if err != nil {
		http.Error(w, "Error creating board", http.StatusInternalServerError)
		return
	}

	log.Printf("Board created with ID: %d", board.ID)

	_, err = app.DB.Exec(`INSERT INTO columns (board_id, status) VALUES
	($1, 'todo'),
	($1, 'doing'),
	($1, 'done')`,
		board.ID)
	if err != nil {
		log.Printf("Error creating columns: %v", err)
		http.Error(w, "Error creating columns", http.StatusInternalServerError)
		return
	}

	response := types.Board{
		ID:        board.ID,
		ProjectID: board.ProjectID,
		Name:      board.Name,
		Columns: []types.Column{
			{
				BoardID: board.ID,
				Status:  "todo",
			},
			{
				BoardID: board.ID,
				Status:  "doing",
			},
			{
				BoardID: board.ID,
				Status:  "done",
			},
		},
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) GetBoardByID(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	var board types.Board

	err := app.DB.QueryRow("SELECT id, project_id, name FROM boards WHERE id = $1", boardID).
		Scan(&board.ID, &board.ProjectID, &board.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

func (app *App) GetBoards(w http.ResponseWriter, r *http.Request) {
	rows, err := app.DB.Query("SELECT id, project_id, name FROM boards")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var boards []types.Board
	for rows.Next() {
		var board types.Board
		if err := rows.Scan(&board.ID, &board.ProjectID, &board.Name); err != nil {
			http.Error(w, "Error scanning boards", http.StatusInternalServerError)
			return
		}
		boards = append(boards, board)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boards)
}

func (app *App) DeleteBoardHandler(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	if boardID == "" {
		http.Error(w, "Board ID is required", http.StatusBadRequest)
		return
	}

	result, err := app.DB.Exec("DELETE FROM boards WHERE id = $1", boardID)
	if err != nil {
		http.Error(w, "Database error while deleting board", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Board not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Board deleted successfully"))
}

func (app *App) UpdateBoardNameHandler(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	if boardID == "" {
		http.Error(w, "Board ID is required", http.StatusBadRequest)
		return
	}

	var updateData types.Board
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingBoard types.Board
	err = app.DB.QueryRow("SELECT id, project_id, name FROM boards WHERE id = $1",
		boardID).Scan(&existingBoard.ID, &existingBoard.ProjectID, &existingBoard.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error while fetching board", http.StatusInternalServerError)
		return
	}

	_, err = app.DB.Exec("UPDATE boards SET name = $1 WHERE id = $2", updateData.Name, boardID)
	if err != nil {
		http.Error(w, "Error updating board name", http.StatusInternalServerError)
		return
	}

	response := types.Board{
		ID:        existingBoard.ID,
		ProjectID: existingBoard.ProjectID,
		Name:      updateData.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
