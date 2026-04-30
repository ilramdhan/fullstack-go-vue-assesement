// 	Usage:
//	make seed         # run migrations and seed (no-op if already seeded)
//	make seed-reset   # delete the sqlite db and re-seed from scratch
package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/config"
	"github.com/durianpay/fullstack-boilerplate/internal/migration"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	_ = godotenv.Load()

	if err := os.MkdirAll(filepath.Dir(config.DatabasePath), 0o755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	dsn := config.DatabasePath + "?_foreign_keys=1&_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := migration.Run(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := migration.Seed(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	var users, payments int
	_ = db.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&users)
	_ = db.QueryRow(`SELECT COUNT(1) FROM payments`).Scan(&payments)
	log.Printf("seeded: users=%d, payments=%d (db=%s)", users, payments, config.DatabasePath)
}
