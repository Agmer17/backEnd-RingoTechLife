package dto

import "mime/multipart"

type SubmitPaymentRequest struct {
	OrderId    string `form:"payment_order_id" validate:"required,uuid"`
	ProofImage *multipart.FileHeader
}

type UpdatePaymentStatusRequest struct {
	PaymentId string  `json:"payment_id" validate:"required,uuid"`
	Note      *string `json:"notes" validate:"omitempty,alphanumspace"`
}
