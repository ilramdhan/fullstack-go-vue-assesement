package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	mw "github.com/durianpay/fullstack-boilerplate/internal/middleware"
	paymentUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

type PaymentHandler struct {
	uc paymentUsecase.PaymentUsecase
}

func NewPaymentHandler(uc paymentUsecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request, params openapigen.ListPaymentsParams) {
	filter := entity.PaymentFilter{}

	if params.Status != nil {
		s := entity.PaymentStatus(*params.Status)
		filter.Status = &s
	}
	if params.Id != nil && *params.Id != "" {
		id := *params.Id
		filter.ID = &id
	}
	if params.Sort != nil {
		filter.Sort = *params.Sort
	}
	if params.Limit != nil {
		filter.Limit = *params.Limit
	}
	if params.Offset != nil {
		filter.Offset = *params.Offset
	}

	items, total, err := h.uc.List(r.Context(), filter)
	if err != nil {
		transport.WriteError(w, err)
		return
	}

	out := openapigen.PaymentList{
		Total: total,
		Data:  make([]openapigen.Payment, 0, len(items)),
	}
	for _, p := range items {
		out.Data = append(out.Data, toAPI(p))
	}
	transport.WriteJSON(w, http.StatusOK, out)
}

func (h *PaymentHandler) GetPaymentSummary(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.Summary(r.Context())
	if err != nil {
		transport.WriteError(w, err)
		return
	}
	transport.WriteJSON(w, http.StatusOK, openapigen.PaymentSummary{
		Total:      s.Total,
		Completed:  s.Completed,
		Processing: s.Processing,
		Failed:     s.Failed,
	})
}

func (h *PaymentHandler) ReviewPayment(w http.ResponseWriter, r *http.Request, id openapigen.PaymentId) {
	var req openapigen.ReviewRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	reviewer := mw.UserEmail(r.Context())
	if reviewer == "" {
		transport.WriteAppError(w, entity.ErrorUnauthorized("missing identity"))
		return
	}

	updated, err := h.uc.Review(r.Context(), id, entity.ReviewDecision(req.Decision), reviewer)
	if err != nil {
		transport.WriteError(w, err)
		return
	}
	transport.WriteJSON(w, http.StatusOK, toAPI(*updated))
}

func toAPI(p entity.Payment) openapigen.Payment {
	return openapigen.Payment{
		Id:         p.ID,
		Merchant:   p.Merchant,
		Amount:     p.Amount,
		Currency:   p.Currency,
		Status:     openapigen.PaymentStatus(p.Status),
		ReviewedBy: p.ReviewedBy,
		ReviewedAt: p.ReviewedAt,
		CreatedAt:  p.CreatedAt,
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil {
		transport.WriteAppError(w, entity.ErrorBadRequest("empty body"))
		return false
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		transport.WriteAppError(w, entity.ErrorBadRequest("failed to read body"))
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			transport.WriteAppError(w, entity.ErrorBadRequest("invalid json"))
			return false
		}
		transport.WriteAppError(w, entity.ErrorBadRequest("invalid json: "+err.Error()))
		return false
	}
	return true
}
