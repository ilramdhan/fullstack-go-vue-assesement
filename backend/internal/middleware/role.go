package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := UserRole(r.Context())
			for _, a := range allowed {
				if role == a {
					next.ServeHTTP(w, r)
					return
				}
			}
			transport.WriteAppError(w, entity.ErrorForbidden(
				fmt.Sprintf("forbidden: requires role %s", strings.Join(quoteAll(allowed), " or ")),
			))
		})
	}
}

func quoteAll(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = fmt.Sprintf("%q", v)
	}
	return out
}
