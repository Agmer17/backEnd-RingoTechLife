package model

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID      uuid.UUID     `json:"id"`
	OrderID uuid.UUID     `json:"order_id"`
	Status  PaymentStatus `json:"status"`
	Amount  float64       `json:"amount"`

	ProofImage *string    `json:"proof_image,omitempty"`
	AdminNote  *string    `json:"admin_note,omitempty"`
	VerifiedBy *uuid.UUID `json:"verified_by,omitempty"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
}
