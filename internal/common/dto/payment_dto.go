package dto

import "mime/multipart"

type SubmitPaymentRequest struct {
	OrderId    string `form:"payment_order_id" validate:"required,uuid"`
	ProofImage *multipart.FileHeader
}
