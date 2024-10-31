package main

import (
	_ "kanban-board/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Example API
// @version 1.0
// @description This is a sample API
// @host localhost:8080
// @BasePath /api

// @Summary Hello World
// @Description Returns a greeting message
// @Produce json
// @Success 200 {string} string "Hello world"
// @Router / [get]
func main() {
	// Создаём новый роутер
	r := gin.Default()

	// Определяем маршрут для корневого сообщения
	r.GET("/api", func(c *gin.Context) {
		c.JSON(200, "Hello world")
	})

	// Запускаем сервер на порту 8080
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080") // устанавливаем :8080 как порт
}
