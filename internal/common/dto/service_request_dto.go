package dto

import "mime/multipart"

// internal/dto/service_request_dto.go

// POST /service-requests — user submit request baru
type CreateServiceRequestDTO struct {
	DeviceType         string  `form:"device_type"          validate:"required,max=100"`
	DeviceBrand        *string `form:"device_brand"         validate:"omitempty,max=100"`
	DeviceModel        *string `form:"device_model"         validate:"omitempty,max=150"`
	ProblemDescription string  `form:"problem_description"  validate:"required"`
	ProductPictures    []*multipart.FileHeader
}

// PATCH /admin/service-requests/:id/quote — admin kasih penawaran
type AdminQuoteServiceRequestDTO struct {
	QuotedPrice       float64 `json:"quoted_price"        validate:"required,min=0"`
	EstimatedDuration int     `json:"estimated_duration"  validate:"required,min=1"`
	AdminNote         *string `json:"admin_note"          validate:"omitempty"`
}

// PATCH /admin/service-requests/:id/reject — admin reject
type AdminRejectServiceRequestDTO struct {
	AdminNote string `json:"admin_note" validate:"required"`
}

// PATCH /service-requests/:id/decision — user acc atau reject penawaran
type UserDecisionDTO struct {
	Accept bool `json:"accept"`
}
