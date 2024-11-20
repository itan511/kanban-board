package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// InitRouter инициализирует роутер с маршрутами
func InitRouter(app *App) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Маршруты для аутентификации
	r.Post("/login", app.Login)
	r.Post("/register", app.RegisterHandler)

	// Маршруты для проектов
	r.Route("/projects", func(r chi.Router) {
		//r.Use(api.TokenVerifyMiddleware) // Защита маршрутов авторизацией
		
		// r.Post("/login", app.Login)
		r.Get("/list", CreateProject)
		// Добавьте остальные CRUD-обработчики для проекта
	})

	// Можно добавить маршруты для колонок и задач, используя аналогичный подход

	return r
}
