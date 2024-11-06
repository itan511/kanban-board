package router

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "kanban-board/api"
)

// InitRouter инициализирует роутер с маршрутами
func InitRouter() *chi.Mux {
    r := chi.NewRouter()
    r.Use(middleware.Logger)

    // Маршруты для аутентификации
    //r.Post("/login", api.Login) // Вход в систему

    // Маршруты для проектов
    r.Route("/projects", func(r chi.Router) {
        //r.Use(api.TokenVerifyMiddleware) // Защита маршрутов авторизацией
        r.Get("/list", api.ListProjects)
        r.Get("/create", api.CreateProject)
        // Добавьте остальные CRUD-обработчики для проекта
    })

    // Можно добавить маршруты для колонок и задач, используя аналогичный подход

    return r
}
