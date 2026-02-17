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
	OrderStatusShipped             OrderStatus = "shipped"
	OrderStatusDelivered           OrderStatus = "delivered"
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

	Subtotal     float64 `json:"subtotal"`
	ShippingCost float64 `json:"shipping_cost"`
	TotalAmount  float64 `json:"total_amount"`

	ShippingName       string `json:"shipping_name"`
	ShippingPhone      string `json:"shipping_phone"`
	ShippingAddress    string `json:"shipping_address"`
	ShippingCity       string `json:"shipping_city"`
	ShippingProvince   string `json:"shipping_province"`
	ShippingPostalCode string `json:"shipping_postal_code"`

	ShippingCourier *string `json:"shipping_courier,omitempty"`
	TrackingNumber  *string `json:"tracking_number,omitempty"`
	Notes           *string `json:"notes,omitempty"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	ShippedAt   *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Relasi — diisi manual saat perlu (bukan auto load)
	Items   []OrderItem `json:"items,omitempty"`
	Payment *Payment    `json:"payment,omitempty"`
}
