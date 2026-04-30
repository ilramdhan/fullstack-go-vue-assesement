package migration

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	Name    string
	Up      string
}

func All() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_users",
			Up: `
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('cs','operation'))
);`,
		},
		{
			Version: 2,
			Name:    "create_payments",
			Up: `
CREATE TABLE IF NOT EXISTS payments (
    id          TEXT PRIMARY KEY,
    merchant    TEXT NOT NULL,
    amount      INTEGER NOT NULL CHECK (amount >= 0),
    currency    TEXT NOT NULL DEFAULT 'IDR',
    status      TEXT NOT NULL CHECK (status IN ('completed','processing','failed')),
    reviewed_by TEXT,
    reviewed_at DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_payments_status     ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);`,
		},
	}
}

func Run(db *sql.DB) error {
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := loadApplied(db)
	if err != nil {
		return err
	}

	for _, m := range All() {
		if applied[m.Version] {
			continue
		}
		if err := apply(db, m); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Name, err)
		}
	}
	return nil
}

func loadApplied(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()

	out := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func apply(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.Up); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations(version, name) VALUES (?, ?)`, m.Version, m.Name); err != nil {
		return err
	}
	return tx.Commit()
}
