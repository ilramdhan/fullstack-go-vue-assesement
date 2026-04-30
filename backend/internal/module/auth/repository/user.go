package repository

import (
	"database/sql"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
)

type UserRepository interface {
	GetUserByEmail(email string) (*entity.User, error)
}

type User struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *User {
	return &User{db: db}
}

func (r *User) GetUserByEmail(email string) (*entity.User, error) {
	row := r.db.QueryRow(`SELECT id, email, password_hash, role FROM users WHERE email = ?`, email)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrorNotFound("user not found")
		}
		return nil, entity.WrapError(err, entity.ErrorCodeInternal, "db error")
	}
	return &u, nil
}
