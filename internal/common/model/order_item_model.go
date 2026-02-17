package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`

	ProductName     string  `json:"product_name"`
	ProductSKU      *string `json:"product_sku,omitempty"`
	PriceAtPurchase float64 `json:"price_at_purchase"`
	Quantity        int     `json:"quantity"`
	Subtotal        float64 `json:"subtotal"`

	CreatedAt time.Time `json:"created_at"`
}
