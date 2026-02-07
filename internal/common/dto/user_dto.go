package dto

import (
	"backEnd-RingoTechLife/internal/common/model"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	FullName    string `json:"full_name" validate:"required,min=3"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required,phoneID"`
	Password    string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	ID                   uuid.UUID `json:"id"`
	FullName             *string   `json:"full_name"`
	Email                *string   `json:"email"`
	Password             *string   `json:"password"`
	PhoneNumber          *string   `json:"phone_number"`
	Role                 *string   `json:"role"`
	ProfilePicture       *string   `json:"profile_picture"`
	DeleteProfilePicture *bool     `json:"delete_profile_picture"`
}

type UserDataResponse struct {
	ID             uuid.UUID `json:"id"`
	FullName       *string   `json:"full_name"`
	Email          *string   `json:"email"`
	PhoneNumber    *string   `json:"phone_number"`
	Role           *string   `json:"role"`
	ProfilePicture *string   `json:"profile_picture"`
}

type DeleteUserRequest struct {
	ID uuid.UUID `json:"id"`
}

func ModelUserToResponse(data model.User) UserDataResponse {
	return UserDataResponse{
		ID:             data.ID,
		FullName:       &data.FullName,
		Email:          &data.Email,
		PhoneNumber:    data.PhoneNumber,
		ProfilePicture: data.ProfilePicture,
		Role:           &data.Role,
	}
}
