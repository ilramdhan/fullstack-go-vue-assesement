package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	authUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxUserEmail
	ctxUserRole
)

func Auth(uc authUsecase.AuthUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				transport.WriteAppError(w, entity.ErrorUnauthorized("missing or invalid authorization header"))
				return
			}
			claims, err := uc.Verify(token)
			if err != nil {
				transport.WriteError(w, err)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxUserEmail, claims.Email)
			ctx = context.WithValue(ctx, ctxUserRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(h[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}

func UserID(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

func UserEmail(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserEmail).(string)
	return v
}

func UserRole(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserRole).(string)
	return v
}
