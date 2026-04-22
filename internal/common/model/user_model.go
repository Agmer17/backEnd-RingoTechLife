package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	PhoneNumber    *string   `json:"phone_number"`
	Password       string    `json:"-"` // jangan expose password
	Role           string    `json:"role"`
	ProfilePicture *string   `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}
