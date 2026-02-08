package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	CreatedAt   time.Time
}
