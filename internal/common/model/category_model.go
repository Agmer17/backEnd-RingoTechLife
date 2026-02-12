package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID `json:"category_id"`
	Name        string    `json:"category_name"`
	Slug        string    `json:"category_slug"`
	Description *string   `json:"category_description"`
	CreatedAt   time.Time `json:"category_created_at"`
}
