package model

import (
	"time"

	"github.com/google/uuid"
)

type ServiceRequestStatus string

const (
	StatusPendingReview   ServiceRequestStatus = "pending_review"
	StatusQuoted          ServiceRequestStatus = "quoted"
	StatusAccepted        ServiceRequestStatus = "accepted"
	StatusRejectedByUser  ServiceRequestStatus = "rejected_by_user"
	StatusRejectedByAdmin ServiceRequestStatus = "rejected_by_admin"
	StatusCancelled       ServiceRequestStatus = "cancelled"
)

type ServiceRequest struct {
	ID                 uuid.UUID            `json:"id"`
	UserID             uuid.UUID            `json:"user_id"`
	DeviceType         string               `json:"device_type"`
	DeviceBrand        *string              `json:"device_brand"`
	DeviceModel        *string              `json:"device_model"`
	ProblemDescription string               `json:"problem_description"`
	Photo1             *string              `json:"photo_1"`
	Photo2             *string              `json:"photo_2"`
	Photo3             *string              `json:"photo_3"`
	Status             ServiceRequestStatus `json:"status"`
	QuotedPrice        *float64             `json:"quoted_price"`
	EstimatedDuration  *int                 `json:"estimated_duration"`
	AdminNote          *string              `json:"admin_note"`
	QuotedBy           *uuid.UUID           `json:"quoted_by"`
	OrderID            *uuid.UUID           `json:"order_id"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	QuotedAt           *time.Time           `json:"quoted_at"`
	DecidedAt          *time.Time           `json:"decided_at"`
}
