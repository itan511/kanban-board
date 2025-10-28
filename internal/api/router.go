package api

import (
	"kanban-board/internal/api/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	chi_mw "github.com/go-chi/chi/v5/middleware"
)

func InitRouter(
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	boardHandler *handlers.BoardHandler,
	columnHandler *handlers.ColumnHandler,
	taskHandler *handlers.TaskHandler,
	authMw func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chi_mw.RequestID)
	r.Use(chi_mw.RealIP)
	r.Use(chi_mw.Logger)
	r.Use(chi_mw.Recoverer)

	r.Post("/login", userHandler.LoginHandler)
	r.Post("/register", userHandler.RegisterHandler)

	r.Group(func(r chi.Router) {
		r.Use(authMw)

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.GetProjectsHandler)
			r.Get("/{id}", projectHandler.GetProjectByIDHandler)
			r.Post("/create", projectHandler.CreateProjectHandler)
			r.Put("/{id}", projectHandler.UpdateProjectHandler)
			r.Delete("/{id}", projectHandler.DeleteProjectHandler)
		})

		r.Route("/project_users", func(r chi.Router) {
			r.Get("/{id}", userHandler.GetProjectUsersHandler)
			r.Post("/add/{id}", userHandler.AddUserToProjectHandler)
			r.Delete("/remove/{id}", userHandler.RemoveUserFromProjectHandler)
		})

		r.Route("/boards", func(r chi.Router) {
			r.Get("/", boardHandler.GetBoardsHandler)
			r.Get("/{id}", boardHandler.GetBoardByIDHandler)
			r.Post("/create", boardHandler.CreateBoardHandler)
			r.Put("/{id}", boardHandler.UpdateBoardHandler)
			r.Delete("/{id}", boardHandler.DeleteBoardHandler)
		})

		r.Route("/columns", func(r chi.Router) {
			r.Get("/", columnHandler.GetColumnsHandler)
			r.Get("/{id}", columnHandler.GetColumnByIDHandler)
			r.Post("/create", columnHandler.CreateColumnHandler)
			r.Put("/{id}", columnHandler.UpdateColumnHandler)
			r.Delete("/{id}", columnHandler.DeleteColumnHandler)
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.GetTasksHandler)
			r.Get("/{id}", taskHandler.GetTaskByIDHandler)
			r.Get("/column/{id}", taskHandler.GetTasksByColumnHandler)
			r.Get("/{id}/logs", taskHandler.GetTaskLogsHandler)
			r.Post("/create", taskHandler.CreateTaskHandler)
			r.Put("/{id}", taskHandler.UpdateTaskHandler)
			r.Delete("/{id}", taskHandler.DeleteTaskHandler)
		})
	})

	return r
}
