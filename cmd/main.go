package main

import (
	"kanban-board/internal/api"
	"kanban-board/internal/api/handlers"
	"kanban-board/internal/db"
	"kanban-board/internal/middleware"
	"kanban-board/internal/repository"
	"kanban-board/internal/services"
	"log"
	"net/http"
	"os"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtKey) == 0 {
		log.Fatal("JWT_SECRET not set in environment variables")
	}

	userRepo := repository.NewUserRepo(database)
	projectRepo := repository.NewProjectRepo(database)
	boardRepo := repository.NewBoardRepo(database)
	columnRepo := repository.NewColumnRepo(database)
	taskRepo := repository.NewTaskRepo(database)

	userSvc := services.NewUserService(userRepo, jwtKey)
	projectSvc := services.NewProjectService(projectRepo)
	boardSvc := services.NewBoardService(boardRepo)
	columnSvc := services.NewColumnService(columnRepo)
	taskSvc := services.NewTaskService(taskRepo)

	userHandler := handlers.NewUserHandler(userSvc)
	projectHandler := handlers.NewProjectHandler(projectSvc)
	boardHandler := handlers.NewBoardHandler(boardSvc)
	columnHandler := handlers.NewColumnHandler(columnSvc)
	taskHandler := handlers.NewTaskHandler(taskSvc)

	authMw := middleware.NewAuthMiddleware(jwtKey)

	r := api.InitRouter(userHandler, projectHandler, boardHandler, columnHandler, taskHandler, authMw)
	log.Println("Routes initialized successfully!")

	log.Println("Starting server on :3000...")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
