package main

import (
    "log"
    "net/http"
    "kanban-board/router"
    //"kanban-board/db"
)

func main() {
    // Подключаемся к базе данных
    //db.InitDB()

    // Инициализируем маршруты
    r := router.InitRouter()

    // Запускаем сервер
    log.Println("Server started at :3000")
    http.ListenAndServe(":3000", r)
}
