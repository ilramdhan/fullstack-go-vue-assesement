package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
)

type stubRepo struct {
	listFn   func(context.Context, entity.PaymentFilter) ([]entity.Payment, int, error)
	getFn    func(context.Context, string) (*entity.Payment, error)
	updateFn func(context.Context, string, entity.PaymentStatus, string) (*entity.Payment, error)
	sumFn    func(context.Context) (entity.PaymentSummary, error)
}

func (s *stubRepo) List(ctx context.Context, f entity.PaymentFilter) ([]entity.Payment, int, error) {
	if s.listFn == nil {
		return nil, 0, nil
	}
	return s.listFn(ctx, f)
}
func (s *stubRepo) GetByID(ctx context.Context, id string) (*entity.Payment, error) {
	return s.getFn(ctx, id)
}
func (s *stubRepo) UpdateReview(ctx context.Context, id string, st entity.PaymentStatus, by string) (*entity.Payment, error) {
	return s.updateFn(ctx, id, st, by)
}
func (s *stubRepo) Summary(ctx context.Context) (entity.PaymentSummary, error) {
	return s.sumFn(ctx)
}

func TestList_RejectsInvalidStatus(t *testing.T) {
	uc := usecase.NewPaymentUsecase(&stubRepo{})
	bogus := entity.PaymentStatus("nope")
	_, _, err := uc.List(context.Background(), entity.PaymentFilter{Status: &bogus})
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeBadRequest {
		t.Fatalf("expected bad_request, got %v", err)
	}
}

func TestList_PassesThroughToRepo(t *testing.T) {
	called := false
	repo := &stubRepo{listFn: func(_ context.Context, f entity.PaymentFilter) ([]entity.Payment, int, error) {
		called = true
		return []entity.Payment{{ID: "p1"}}, 1, nil
	}}
	uc := usecase.NewPaymentUsecase(repo)
	out, total, err := uc.List(context.Background(), entity.PaymentFilter{})
	if err != nil || !called || total != 1 || len(out) != 1 {
		t.Fatalf("unexpected: err=%v called=%v total=%d out=%v", err, called, total, out)
	}
}

func TestReview_RejectsInvalidDecision(t *testing.T) {
	uc := usecase.NewPaymentUsecase(&stubRepo{})
	_, err := uc.Review(context.Background(), "p1", entity.ReviewDecision("maybe"), "x@test.com")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeBadRequest {
		t.Fatalf("expected bad_request, got %v", err)
	}
}

func TestReview_NotFound(t *testing.T) {
	repo := &stubRepo{getFn: func(_ context.Context, _ string) (*entity.Payment, error) {
		return nil, entity.ErrorNotFound("payment not found")
	}}
	uc := usecase.NewPaymentUsecase(repo)
	_, err := uc.Review(context.Background(), "missing", entity.ReviewApprove, "x@test.com")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

func TestReview_ConflictWhenNotProcessing(t *testing.T) {
	repo := &stubRepo{getFn: func(_ context.Context, _ string) (*entity.Payment, error) {
		return &entity.Payment{ID: "p1", Status: entity.PaymentStatusCompleted}, nil
	}}
	uc := usecase.NewPaymentUsecase(repo)
	_, err := uc.Review(context.Background(), "p1", entity.ReviewApprove, "x@test.com")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestReview_ApproveTransitionsToCompleted(t *testing.T) {
	now := time.Now().UTC()
	gotStatus := entity.PaymentStatus("")
	gotReviewer := ""
	repo := &stubRepo{
		getFn: func(_ context.Context, _ string) (*entity.Payment, error) {
			return &entity.Payment{ID: "p1", Status: entity.PaymentStatusProcessing}, nil
		},
		updateFn: func(_ context.Context, id string, s entity.PaymentStatus, by string) (*entity.Payment, error) {
			gotStatus = s
			gotReviewer = by
			return &entity.Payment{ID: id, Status: s, ReviewedBy: &by, ReviewedAt: &now}, nil
		},
	}
	uc := usecase.NewPaymentUsecase(repo)
	out, err := uc.Review(context.Background(), "p1", entity.ReviewApprove, "op@test.com")
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if gotStatus != entity.PaymentStatusCompleted {
		t.Errorf("status: got %s want completed", gotStatus)
	}
	if gotReviewer != "op@test.com" {
		t.Errorf("reviewer: got %s want op@test.com", gotReviewer)
	}
	if out.Status != entity.PaymentStatusCompleted {
		t.Errorf("returned status: got %s", out.Status)
	}
}

func TestReview_RejectTransitionsToFailed(t *testing.T) {
	gotStatus := entity.PaymentStatus("")
	repo := &stubRepo{
		getFn: func(_ context.Context, _ string) (*entity.Payment, error) {
			return &entity.Payment{ID: "p1", Status: entity.PaymentStatusProcessing}, nil
		},
		updateFn: func(_ context.Context, id string, s entity.PaymentStatus, by string) (*entity.Payment, error) {
			gotStatus = s
			return &entity.Payment{ID: id, Status: s, ReviewedBy: &by}, nil
		},
	}
	uc := usecase.NewPaymentUsecase(repo)
	_, err := uc.Review(context.Background(), "p1", entity.ReviewReject, "op@test.com")
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if gotStatus != entity.PaymentStatusFailed {
		t.Errorf("status: got %s want failed", gotStatus)
	}
}

func TestSummary_PassesThrough(t *testing.T) {
	repo := &stubRepo{sumFn: func(_ context.Context) (entity.PaymentSummary, error) {
		return entity.PaymentSummary{Total: 5}, nil
	}}
	uc := usecase.NewPaymentUsecase(repo)
	got, err := uc.Summary(context.Background())
	if err != nil || got.Total != 5 {
		t.Fatalf("Summary: err=%v got=%+v", err, got)
	}
}
