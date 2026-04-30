package repository_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/migration"
	"github.com/durianpay/fullstack-boilerplate/internal/module/auth/repository"
	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migration.Run(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestGetUserByEmail_Found(t *testing.T) {
	db := newTestDB(t)
	if _, err := db.Exec(`INSERT INTO users(email, password_hash, role) VALUES (?, ?, ?)`,
		"a@test.com", "hash", "cs"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := repository.NewUserRepo(db)
	u, err := r.GetUserByEmail("a@test.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if u.Email != "a@test.com" || u.Role != "cs" {
		t.Errorf("got %+v", u)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	db := newTestDB(t)
	r := repository.NewUserRepo(db)
	_, err := r.GetUserByEmail("missing@test.com")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}
