package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
)

type PaymentRepository interface {
	List(ctx context.Context, f entity.PaymentFilter) ([]entity.Payment, int, error)
	GetByID(ctx context.Context, id string) (*entity.Payment, error)
	UpdateReview(ctx context.Context, id string, status entity.PaymentStatus, reviewer string) (*entity.Payment, error)
	Summary(ctx context.Context) (entity.PaymentSummary, error)
}

type Payment struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *Payment { return &Payment{db: db} }

// allowedSortColumns whitelists `sort` query values to prevent SQL injection.
var allowedSortColumns = map[string]string{
	"created_at":  "created_at ASC",
	"-created_at": "created_at DESC",
	"amount":      "amount ASC",
	"-amount":     "amount DESC",
}

func (r *Payment) List(ctx context.Context, f entity.PaymentFilter) ([]entity.Payment, int, error) {
	where := []string{"1=1"}
	args := []any{}

	if f.Status != nil {
		where = append(where, "status = ?")
		args = append(args, string(*f.Status))
	}
	if f.ID != nil && *f.ID != "" {
		where = append(where, "id = ?")
		args = append(args, *f.ID)
	}

	orderBy, ok := allowedSortColumns[f.Sort]
	if !ok {
		orderBy = "created_at DESC"
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	countQ := fmt.Sprintf(`SELECT COUNT(1) FROM payments WHERE %s`, strings.Join(where, " AND "))
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, entity.WrapError(err, entity.ErrorCodeInternal, "count payments")
	}

	q := fmt.Sprintf(`
SELECT id, merchant, amount, currency, status, reviewed_by, reviewed_at, created_at
FROM payments
WHERE %s
ORDER BY %s
LIMIT ? OFFSET ?`, strings.Join(where, " AND "), orderBy)
	args = append(args, limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, entity.WrapError(err, entity.ErrorCodeInternal, "list payments")
	}
	defer rows.Close()

	out := make([]entity.Payment, 0, limit)
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, 0, entity.WrapError(err, entity.ErrorCodeInternal, "scan payment")
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func (r *Payment) GetByID(ctx context.Context, id string) (*entity.Payment, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, merchant, amount, currency, status, reviewed_by, reviewed_at, created_at
FROM payments WHERE id = ?`, id)

	p, err := scanPayment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrorNotFound("payment not found")
		}
		return nil, entity.WrapError(err, entity.ErrorCodeInternal, "get payment")
	}
	return &p, nil
}

func (r *Payment) UpdateReview(ctx context.Context, id string, status entity.PaymentStatus, reviewer string) (*entity.Payment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, `
UPDATE payments
   SET status = ?, reviewed_by = ?, reviewed_at = ?
 WHERE id = ?`, string(status), reviewer, now, id)
	if err != nil {
		return nil, entity.WrapError(err, entity.ErrorCodeInternal, "update payment")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, entity.ErrorNotFound("payment not found")
	}
	return r.GetByID(ctx, id)
}

func (r *Payment) Summary(ctx context.Context) (entity.PaymentSummary, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(1) FROM payments GROUP BY status`)
	if err != nil {
		return entity.PaymentSummary{}, entity.WrapError(err, entity.ErrorCodeInternal, "summary")
	}
	defer rows.Close()

	var s entity.PaymentSummary
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return entity.PaymentSummary{}, entity.WrapError(err, entity.ErrorCodeInternal, "scan summary")
		}
		switch entity.PaymentStatus(status) {
		case entity.PaymentStatusCompleted:
			s.Completed = n
		case entity.PaymentStatusProcessing:
			s.Processing = n
		case entity.PaymentStatusFailed:
			s.Failed = n
		}
	}
	s.Total = s.Completed + s.Processing + s.Failed
	return s, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPayment(s rowScanner) (entity.Payment, error) {
	var p entity.Payment
	var status string
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime
	var createdAt time.Time

	if err := s.Scan(
		&p.ID, &p.Merchant, &p.Amount, &p.Currency,
		&status, &reviewedBy, &reviewedAt, &createdAt,
	); err != nil {
		return entity.Payment{}, err
	}
	p.Status = entity.PaymentStatus(status)
	p.CreatedAt = createdAt
	if reviewedBy.Valid {
		p.ReviewedBy = &reviewedBy.String
	}
	if reviewedAt.Valid {
		p.ReviewedAt = &reviewedAt.Time
	}
	return p, nil
}
