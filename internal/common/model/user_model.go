package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	FullName       string
	Email          string
	PhoneNumber    *string
	Password       string
	Role           string
	ProfilePicture *string
	CreatedAt      time.Time
}
