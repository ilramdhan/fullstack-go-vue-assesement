package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/api"
	"github.com/durianpay/fullstack-boilerplate/internal/config"
	ah "github.com/durianpay/fullstack-boilerplate/internal/module/auth/handler"
	ar "github.com/durianpay/fullstack-boilerplate/internal/module/auth/repository"
	au "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	srv "github.com/durianpay/fullstack-boilerplate/internal/service/http"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()

	if err := os.MkdirAll(filepath.Dir(config.DatabasePath), 0o755); err != nil {
		log.Fatalf("create db dir: %v", err)
	}

	dsn := config.DatabasePath + "?_foreign_keys=1&_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := initDB(db); err != nil {
		log.Fatal(err)
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

func initDB(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
		  id INTEGER PRIMARY KEY AUTOINCREMENT,
		  email TEXT NOT NULL UNIQUE,
		  password_hash TEXT NOT NULL,
		  role TEXT NOT NULL
		);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	var cnt int
	if err := db.QueryRow("SELECT COUNT(1) FROM users").Scan(&cnt); err != nil {
		return err
	}
	if cnt == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		seeds := []struct{ email, role string }{
			{"cs@test.com", "cs"},
			{"operation@test.com", "operation"},
		}
		for _, s := range seeds {
			if _, err := db.Exec("INSERT INTO users(email, password_hash, role) VALUES (?, ?, ?)",
				s.email, string(hash), s.role); err != nil {
				return err
			}
		}
	}
	return nil
}
