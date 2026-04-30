package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/migration"
	"github.com/durianpay/fullstack-boilerplate/internal/module/payment/repository"
	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migration.Run(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func insert(t *testing.T, db *sql.DB, id, merchant string, amount int64, status entity.PaymentStatus, createdAt time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO payments(id, merchant, amount, currency, status, created_at) VALUES (?, ?, ?, 'IDR', ?, ?)`,
		id, merchant, amount, string(status), createdAt.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
}

func TestList_FilterByStatus(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().UTC()
	insert(t, db, "p1", "Tokopedia", 100, entity.PaymentStatusCompleted, now)
	insert(t, db, "p2", "Shopee", 200, entity.PaymentStatusProcessing, now)
	insert(t, db, "p3", "Lazada", 300, entity.PaymentStatusFailed, now)
	insert(t, db, "p4", "Blibli", 400, entity.PaymentStatusProcessing, now)

	r := repository.NewPaymentRepo(db)
	status := entity.PaymentStatusProcessing
	got, total, err := r.List(context.Background(), entity.PaymentFilter{Status: &status})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 || len(got) != 2 {
		t.Fatalf("want 2 processing, got total=%d len=%d", total, len(got))
	}
	for _, p := range got {
		if p.Status != entity.PaymentStatusProcessing {
			t.Errorf("unexpected status %s", p.Status)
		}
	}
}

func TestList_FilterByID(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().UTC()
	insert(t, db, "p1", "Tokopedia", 100, entity.PaymentStatusCompleted, now)
	insert(t, db, "p2", "Shopee", 200, entity.PaymentStatusProcessing, now)

	r := repository.NewPaymentRepo(db)
	id := "p2"
	got, total, err := r.List(context.Background(), entity.PaymentFilter{ID: &id})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(got) != 1 || got[0].ID != "p2" {
		t.Fatalf("want only p2, got total=%d items=%v", total, got)
	}
}

func TestList_SortAndPagination(t *testing.T) {
	db := newTestDB(t)
	base := time.Now().UTC().Add(-24 * time.Hour)
	for i, amount := range []int64{500, 100, 300, 200, 400} {
		insert(t, db, "p"+string(rune('a'+i)), "M", amount, entity.PaymentStatusCompleted, base.Add(time.Duration(i)*time.Hour))
	}

	r := repository.NewPaymentRepo(db)

	got, _, err := r.List(context.Background(), entity.PaymentFilter{Sort: "amount"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 5 || got[0].Amount != 100 || got[4].Amount != 500 {
		t.Fatalf("ascending sort by amount failed: %+v", amounts(got))
	}

	got, _, err = r.List(context.Background(), entity.PaymentFilter{Sort: "-amount"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got[0].Amount != 500 {
		t.Fatalf("descending sort by amount failed: %+v", amounts(got))
	}

	got, total, err := r.List(context.Background(), entity.PaymentFilter{Limit: 2, Offset: 2, Sort: "amount"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 5 || len(got) != 2 || got[0].Amount != 300 {
		t.Fatalf("pagination failed: total=%d items=%v", total, amounts(got))
	}
}

func TestList_UnknownSortFallsBackToCreatedAtDesc(t *testing.T) {
	db := newTestDB(t)
	base := time.Now().UTC().Add(-24 * time.Hour)
	insert(t, db, "old", "M", 100, entity.PaymentStatusCompleted, base)
	insert(t, db, "new", "M", 100, entity.PaymentStatusCompleted, base.Add(2*time.Hour))

	r := repository.NewPaymentRepo(db)
	got, _, err := r.List(context.Background(), entity.PaymentFilter{Sort: "; DROP TABLE payments;"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 || got[0].ID != "new" {
		t.Fatalf("expected newest first, got %+v", got)
	}
}

func TestList_EmptyResult(t *testing.T) {
	db := newTestDB(t)
	r := repository.NewPaymentRepo(db)
	got, total, err := r.List(context.Background(), entity.PaymentFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 0 || len(got) != 0 {
		t.Fatalf("expected empty, got total=%d items=%v", total, got)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	r := repository.NewPaymentRepo(db)
	_, err := r.GetByID(context.Background(), "missing")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeNotFound {
		t.Fatalf("expected not_found AppError, got %v", err)
	}
}

func TestUpdateReview_SetsReviewerAndTimestamp(t *testing.T) {
	db := newTestDB(t)
	insert(t, db, "p1", "Tokopedia", 100, entity.PaymentStatusProcessing, time.Now().UTC())

	r := repository.NewPaymentRepo(db)
	got, err := r.UpdateReview(context.Background(), "p1", entity.PaymentStatusCompleted, "operation@test.com")
	if err != nil {
		t.Fatalf("UpdateReview: %v", err)
	}
	if got.Status != entity.PaymentStatusCompleted {
		t.Errorf("status: got %s want completed", got.Status)
	}
	if got.ReviewedBy == nil || *got.ReviewedBy != "operation@test.com" {
		t.Errorf("reviewed_by: got %+v", got.ReviewedBy)
	}
	if got.ReviewedAt == nil {
		t.Errorf("reviewed_at: got nil")
	}
}

func TestUpdateReview_NotFound(t *testing.T) {
	db := newTestDB(t)
	r := repository.NewPaymentRepo(db)
	_, err := r.UpdateReview(context.Background(), "missing", entity.PaymentStatusCompleted, "x")
	var appErr *entity.AppError
	if !errors.As(err, &appErr) || appErr.Code != entity.ErrorCodeNotFound {
		t.Fatalf("expected not_found, got %v", err)
	}
}

func TestSummary(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		insert(t, db, "c"+string(rune('a'+i)), "M", 100, entity.PaymentStatusCompleted, now)
	}
	for i := 0; i < 2; i++ {
		insert(t, db, "p"+string(rune('a'+i)), "M", 100, entity.PaymentStatusProcessing, now)
	}
	insert(t, db, "f1", "M", 100, entity.PaymentStatusFailed, now)

	r := repository.NewPaymentRepo(db)
	s, err := r.Summary(context.Background())
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if s.Completed != 3 || s.Processing != 2 || s.Failed != 1 || s.Total != 6 {
		t.Errorf("summary mismatch: %+v", s)
	}
}

func TestSummary_EmptyTable(t *testing.T) {
	db := newTestDB(t)
	r := repository.NewPaymentRepo(db)
	s, err := r.Summary(context.Background())
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if s.Total != 0 {
		t.Errorf("expected zero summary, got %+v", s)
	}
}

func amounts(ps []entity.Payment) []int64 {
	out := make([]int64, len(ps))
	for i, p := range ps {
		out[i] = p.Amount
	}
	return out
}
