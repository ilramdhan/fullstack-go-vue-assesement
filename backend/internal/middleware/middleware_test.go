package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/middleware"
	"github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
)

type stubAuth struct {
	verifyFn func(string) (*usecase.Claims, error)
}

func (s *stubAuth) Login(_, _ string) (string, *entity.User, error) {
	return "", nil, nil
}

func (s *stubAuth) Verify(token string) (*usecase.Claims, error) {
	return s.verifyFn(token)
}

func TestAuth_MissingHeader(t *testing.T) {
	mw := middleware.Auth(&stubAuth{})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rec.Code)
	}
}

func TestAuth_NotBearer(t *testing.T) {
	mw := middleware.Auth(&stubAuth{})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Basic abc")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rec.Code)
	}
}

func TestAuth_EmptyBearer(t *testing.T) {
	mw := middleware.Auth(&stubAuth{})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer    ")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rec.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	mw := middleware.Auth(&stubAuth{
		verifyFn: func(string) (*usecase.Claims, error) {
			return nil, entity.ErrorUnauthorized("bad token")
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer junk")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want 401", rec.Code)
	}
}

func TestAuth_ValidTokenInjectsContext(t *testing.T) {
	mw := middleware.Auth(&stubAuth{
		verifyFn: func(string) (*usecase.Claims, error) {
			return &usecase.Claims{UserID: "1", Email: "u@test.com", Role: "operation"}, nil
		},
	})
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := middleware.UserID(r.Context()); got != "1" {
			t.Errorf("UserID: got %q", got)
		}
		if got := middleware.UserEmail(r.Context()); got != "u@test.com" {
			t.Errorf("UserEmail: got %q", got)
		}
		if got := middleware.UserRole(r.Context()); got != "operation" {
			t.Errorf("UserRole: got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer goodtoken")
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)

	if !called {
		t.Fatalf("next handler not invoked")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d want 200", rec.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	auth := middleware.Auth(&stubAuth{
		verifyFn: func(string) (*usecase.Claims, error) {
			return &usecase.Claims{UserID: "1", Role: "operation"}, nil
		},
	})
	guard := middleware.RequireRole("operation")
	called := false
	chain := auth(guard(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusOK {
		t.Fatalf("operation should pass: code=%d called=%v", rec.Code, called)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	auth := middleware.Auth(&stubAuth{
		verifyFn: func(string) (*usecase.Claims, error) {
			return &usecase.Claims{UserID: "1", Role: "cs"}, nil
		},
	})
	guard := middleware.RequireRole("operation")
	chain := auth(guard(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	})))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer t")
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got %d want 403", rec.Code)
	}
}
