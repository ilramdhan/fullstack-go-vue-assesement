package api

import (
	"net/http"

	ah "github.com/durianpay/fullstack-boilerplate/internal/module/auth/handler"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

// APIHandler is a thin adapter implementing the generated ServerInterface.
// Each method delegates to a per-module handler.
type APIHandler struct {
	Auth *ah.AuthHandler
}

// Compile-time guarantee: APIHandler must implement every operation in openapi.yaml.
var _ openapigen.ServerInterface = (*APIHandler)(nil)

func (h *APIHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	h.Auth.LoginUser(w, r)
}

func (h *APIHandler) ListPayments(w http.ResponseWriter, _ *http.Request, _ openapigen.ListPaymentsParams) {
	notImplemented(w)
}

func (h *APIHandler) GetPaymentSummary(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

func (h *APIHandler) ReviewPayment(w http.ResponseWriter, _ *http.Request, _ openapigen.PaymentId) {
	notImplemented(w)
}

func notImplemented(w http.ResponseWriter) {
	transport.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"code":    http.StatusNotImplemented,
		"message": "endpoint not implemented yet",
	})
}
