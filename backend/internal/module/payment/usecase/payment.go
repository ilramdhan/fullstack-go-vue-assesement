package usecase

import (
	"context"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/payment/repository"
)

type PaymentUsecase interface {
	List(ctx context.Context, f entity.PaymentFilter) ([]entity.Payment, int, error)
	Summary(ctx context.Context) (entity.PaymentSummary, error)
	Review(ctx context.Context, id string, decision entity.ReviewDecision, reviewer string) (*entity.Payment, error)
}

type Payment struct {
	repo repository.PaymentRepository
}

func NewPaymentUsecase(repo repository.PaymentRepository) *Payment {
	return &Payment{repo: repo}
}

func (u *Payment) List(ctx context.Context, f entity.PaymentFilter) ([]entity.Payment, int, error) {
	if f.Status != nil && !f.Status.Valid() {
		return nil, 0, entity.ErrorBadRequest("invalid status filter")
	}
	return u.repo.List(ctx, f)
}

func (u *Payment) Summary(ctx context.Context) (entity.PaymentSummary, error) {
	return u.repo.Summary(ctx)
}

func (u *Payment) Review(ctx context.Context, id string, decision entity.ReviewDecision, reviewer string) (*entity.Payment, error) {
	if !decision.Valid() {
		return nil, entity.ErrorBadRequest("decision must be 'approve' or 'reject'")
	}
	current, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Status != entity.PaymentStatusProcessing {
		return nil, entity.ErrorConflict(
			"payment is " + string(current.Status) + " and cannot be reviewed",
		)
	}
	return u.repo.UpdateReview(ctx, id, decision.Resolve(), reviewer)
}
