package dto

import "github.com/google/uuid"

type CreateUserRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type UpdateUserRequest struct {
	ID             uuid.UUID `json:"id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	PhoneNumber    string    `json:"phone_number"`
	Role           string    `json:"role"`
	ProfilePicture *string   `json:"profile_picture"`
}

type DeleteUserRequest struct {
	ID uuid.UUID `json:"id"`
}
