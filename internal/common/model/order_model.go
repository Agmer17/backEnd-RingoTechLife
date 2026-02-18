package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string
type PaymentStatus string

const (
	OrderStatusPending             OrderStatus = "pending"
	OrderStatusWaitingConfirmation OrderStatus = "waiting_confirmation"
	OrderStatusConfirmed           OrderStatus = "confirmed"
	OrderStatusCancelled           OrderStatus = "cancelled"
)

const (
	PaymentStatusUnpaid    PaymentStatus = "unpaid"
	PaymentStatusSubmitted PaymentStatus = "submitted"
	PaymentStatusApproved  PaymentStatus = "approved"
	PaymentStatusRejected  PaymentStatus = "rejected"
)

type Order struct {
	ID     uuid.UUID   `json:"id"`
	UserID uuid.UUID   `json:"user_id"`
	Status OrderStatus `json:"status"`

	Subtotal    float64 `json:"subtotal"`
	TotalAmount float64 `json:"total_amount"`
	Notes       *string `json:"notes,omitempty"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Relasi — diisi manual saat perlu (bukan auto load)
	Items   []OrderItem `json:"items,omitempty"`
	Payment *Payment    `json:"payment,omitempty"`
}
