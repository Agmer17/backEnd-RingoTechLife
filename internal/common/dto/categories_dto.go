package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCategoryRequest struct {
	Name string  `json:"category_name" validate:"required,min=2,max=100"`
	Slug string  `json:"category_slug" validate:"required,slug,max=100"`
	Desc *string `json:"category_description" validate:"omitempty,max=255"`
}

type UpdateCategoryRequest struct {
	Name *string `json:"category_name" validate:"omitempty,min=2,max=100"`
	Slug *string `json:"category_slug" validate:"omitempty,slug,max=100"`
	Desc *string `json:"category_description" validate:"omitempty,max=255"`
}

type DetailCategoryResponse struct {
	ID           uuid.UUID `json:"category_id"`
	Name         string    `json:"category_name"`
	Slug         string    `json:"category_slug"`
	Description  *string   `json:"category_description"`
	ProductCount int       `json:"total_product"`
	CreatedAt    time.Time `json:"category_created_at"`
}
