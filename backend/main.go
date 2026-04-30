package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/api"
	"github.com/durianpay/fullstack-boilerplate/internal/config"
	"github.com/durianpay/fullstack-boilerplate/internal/migration"
	ah "github.com/durianpay/fullstack-boilerplate/internal/module/auth/handler"
	ar "github.com/durianpay/fullstack-boilerplate/internal/module/auth/repository"
	au "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	srv "github.com/durianpay/fullstack-boilerplate/internal/service/http"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	_ = godotenv.Load()

	db, err := openDB(config.DatabasePath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migration.Run(db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	if err := migration.Seed(db); err != nil {
		log.Fatalf("seed db: %v", err)
	}

	jwtExpiry, err := time.ParseDuration(config.JwtExpired)
	if err != nil {
		log.Fatalf("invalid JWT_EXPIRED: %v", err)
	}

	userRepo := ar.NewUserRepo(db)
	authUC := au.NewAuthUsecase(userRepo, config.JwtSecret, jwtExpiry)
	authH := ah.NewAuthHandler(authUC)

	apiHandler := &api.APIHandler{
		Auth: authH,
	}

	server := srv.NewServer(apiHandler, config.OpenapiYamlLocation)
	log.Printf("starting server on %s (db=%s)", config.HttpAddress, config.DatabasePath)
	server.Start(config.HttpAddress)
}

// openDB opens the sqlite database with WAL journaling and a tuned pool.
func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	dsn := path + "?_foreign_keys=1&_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}
