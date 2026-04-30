package entity

import "time"

type PaymentStatus string

const (
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusFailed     PaymentStatus = "failed"
)

func (s PaymentStatus) Valid() bool {
	switch s {
	case PaymentStatusCompleted, PaymentStatusProcessing, PaymentStatusFailed:
		return true
	}
	return false
}

type ReviewDecision string

const (
	ReviewApprove ReviewDecision = "approve"
	ReviewReject  ReviewDecision = "reject"
)

func (d ReviewDecision) Valid() bool {
	return d == ReviewApprove || d == ReviewReject
}

func (d ReviewDecision) Resolve() PaymentStatus {
	if d == ReviewApprove {
		return PaymentStatusCompleted
	}
	return PaymentStatusFailed
}

type Payment struct {
	ID         string
	Merchant   string
	Amount     int64
	Currency   string
	Status     PaymentStatus
	ReviewedBy *string
	ReviewedAt *time.Time
	CreatedAt  time.Time
}

type PaymentFilter struct {
	Status *PaymentStatus
	ID     *string
	Sort   string
	Limit  int
	Offset int
}

type PaymentSummary struct {
	Total      int
	Completed  int
	Processing int
	Failed     int
}
