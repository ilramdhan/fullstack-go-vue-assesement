package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"golang.org/x/crypto/bcrypt"
)

type stubRepo struct {
	user *entity.User
	err  error
}

func (s *stubRepo) GetUserByEmail(_ string) (*entity.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func newUC(repo *stubRepo, ttl time.Duration) *usecase.Auth {
	return usecase.NewAuthUsecase(repo, []byte("test-secret-test-secret"), ttl)
}

func hashedUser(t *testing.T, password string) *entity.User {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return &entity.User{ID: "1", Email: "u@test.com", PasswordHash: string(h), Role: "cs"}
}

func TestLogin_Success(t *testing.T) {
	uc := newUC(&stubRepo{user: hashedUser(t, "secret")}, time.Hour)
	tok, u, err := uc.Login("u@test.com", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if tok == "" || u.Email != "u@test.com" {
		t.Errorf("unexpected token=%q user=%+v", tok, u)
	}

	claims, err := uc.Verify(tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != "1" || claims.Email != "u@test.com" || claims.Role != "cs" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	uc := newUC(&stubRepo{user: hashedUser(t, "secret")}, time.Hour)
	_, _, err := uc.Login("u@test.com", "WRONG")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestLogin_RepoError(t *testing.T) {
	uc := newUC(&stubRepo{err: entity.ErrorNotFound("nope")}, time.Hour)
	_, _, err := uc.Login("u@test.com", "x")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeNotFound {
		t.Fatalf("expected not_found from repo, got %v", err)
	}
}

func TestVerify_InvalidToken(t *testing.T) {
	uc := newUC(&stubRepo{}, time.Hour)
	_, err := uc.Verify("not.a.token")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	uc := newUC(&stubRepo{user: hashedUser(t, "secret")}, -time.Second)
	tok, _, err := uc.Login("u@test.com", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	_, err = uc.Verify(tok)
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeUnauthorized {
		t.Fatalf("expected unauthorized for expired, got %v", err)
	}
}

func TestVerify_DifferentSecretRejected(t *testing.T) {
	repo := &stubRepo{user: hashedUser(t, "secret")}
	signer := newUC(repo, time.Hour)
	tok, _, _ := signer.Login("u@test.com", "secret")

	other := usecase.NewAuthUsecase(repo, []byte("different-secret-different-secret"), time.Hour)
	if _, err := other.Verify(tok); err == nil {
		t.Fatalf("expected rejection on different secret")
	}
}
