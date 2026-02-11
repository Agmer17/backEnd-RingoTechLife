package dto

import (
	"backEnd-RingoTechLife/internal/common/model"
	"mime/multipart"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	FullName    string `json:"full_name" validate:"required,min=3"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required,phoneID"`
	Password    string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	ID                   uuid.UUID `validate:"omitempty,uuid" form:"id"`
	FullName             *string   `validate:"omitempty,min=4,aplhanumspace" form:"full_name"`
	Email                *string   `validate:"omitempty,email" form:"email"`
	Password             *string   `validate:"omitempty,min=8" form:"password"`
	PhoneNumber          *string   `validate:"omitempty,phoneID" form:"phone_number"`
	Role                 *string   `validate:"omitempty" form:"role"`
	ProfilePicture       *multipart.FileHeader
	DeleteProfilePicture *bool `validate:"omitempty" form:"delete_profile_picture"`
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
