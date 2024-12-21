package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	err, dbHost := loadEnv()
	if err != nil {
		return nil, err
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSslMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbSslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	
	if err := waitForDB(db); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Connected to the database successfully!")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Error initializing postgres driver: ", err)
		return nil, err
	}

	migrationPath := getMigrationPath()

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"postgres", driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied successfully!")
	return db, nil
}

func loadEnv() (error, string) {
	var dbHost string

	if _, err := os.Stat("/app/.env"); os.IsNotExist(err) {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file from root: %v", err)
			return err, ""
		}
		dbHost = os.Getenv("DB_DEV_HOST")
	} else {
		err := godotenv.Load("/app/.env")
		if err != nil {
			log.Fatalf("Error loading .env file from /app: %v", err)
			return err, ""
		}
		dbHost = os.Getenv("DB_HOST")
	}

	if dbHost == "" {
		return fmt.Errorf("DB_HOST not set in .env file"), ""
	}

	return nil, dbHost
}

func getMigrationPath() string {
	if _, err := os.Stat("/app"); os.IsNotExist(err) {
		return "file://./migrations"
	} else {
		return "file:///app/migrations"
	}
}

func waitForDB(db *sql.DB) error {
	for {
		if err := db.Ping(); err == nil {
			return nil
		}
		log.Println("Waiting for database...")
		time.Sleep(2 * time.Second)
	}
}
