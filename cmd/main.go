package main

import (
	"kanban-board/api"
	"kanban-board/db"
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

	app := &api.App{
		DB:     database,
		JWTKey: jwtKey,
	}

	r := api.InitRouter(app)
	log.Println("Routes initialized successfully!")

	log.Println("Starting server...")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
