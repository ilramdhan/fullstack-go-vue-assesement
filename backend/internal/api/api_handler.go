package api

import (
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	mw "github.com/durianpay/fullstack-boilerplate/internal/middleware"
	ah "github.com/durianpay/fullstack-boilerplate/internal/module/auth/handler"
	ph "github.com/durianpay/fullstack-boilerplate/internal/module/payment/handler"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

type APIHandler struct {
	Auth    *ah.AuthHandler
	Payment *ph.PaymentHandler
}

var _ openapigen.ServerInterface = (*APIHandler)(nil)

func (h *APIHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	h.Auth.LoginUser(w, r)
}

func (h *APIHandler) ListPayments(w http.ResponseWriter, r *http.Request, params openapigen.ListPaymentsParams) {
	h.Payment.ListPayments(w, r, params)
}

func (h *APIHandler) GetPaymentSummary(w http.ResponseWriter, r *http.Request) {
	h.Payment.GetPaymentSummary(w, r)
}

func (h *APIHandler) ReviewPayment(w http.ResponseWriter, r *http.Request, id openapigen.PaymentId) {
	if mw.UserRole(r.Context()) != "operation" {
		transport.WriteAppError(w, entity.ErrorForbidden(`forbidden: requires role "operation"`))
		return
	}
	h.Payment.ReviewPayment(w, r, id)
}
