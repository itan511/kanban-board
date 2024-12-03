package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(app *App) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/login", app.LoginHandler)
	r.Post("/register", app.RegisterHandler)

	r.Route("/projects", func(r chi.Router) {
		r.Get("/", app.GetProjects)
		r.Get("/{id}", app.GetProjectByID)
		r.Post("/create", app.CreateProjectHandler)
		r.Put("/{id}", app.UpdateProjectHandler)
		r.Delete("/{id}", app.DeleteProjectHandler)
	})

	r.Route("/boards", func(r chi.Router) {
		r.Get("/", app.GetBoards)
		r.Get("/{id}", app.GetBoardByID)
		r.Post("/create", app.CreateBoardHandler)
		r.Put("/{id}", app.UpdateBoardNameHandler)
		r.Delete("/{id}", app.DeleteBoardHandler)
	})

	r.Route("/project_users", func(r chi.Router) {
		r.Get("/{id}", app.GetProjectUsersHandler)
		r.Post("/add/{id}", app.AddUserToProjectHandler)
		r.Delete("/remove/{id}", app.RemoveUserFromProjectHandler)
	})

	return r
}
