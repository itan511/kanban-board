package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Для миграций через файлы
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Для драйвера PostgreSQL
)

func InitDB() (*sql.DB, error) {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return nil, err
	}

	// Получаем переменные окружения для подключения к базе данных
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSslMode := os.Getenv("DB_SSLMODE")

	// Формируем строку подключения
	connStr := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=" + dbSslMode

	// Создаем подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	log.Println("Connected to the database successfully!")

	// Инициализируем миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Error initializing postgres driver: ", err)
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	m.Up()

	log.Println("Migrations applied successfully!")
	return db, nil
}
