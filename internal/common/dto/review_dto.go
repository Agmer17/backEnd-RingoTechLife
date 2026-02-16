package dto

import (
	"time"

	"github.com/google/uuid"
)

// ReviewUser hanya expose field user yang aman untuk ditampilkan ke client.
type ReviewUser struct {
	ID             uuid.UUID `json:"user_id"`
	FullName       string    `json:"user_fullname"`
	ProfilePicture *string   `json:"user_profile_picture"`
}

// ReviewDetail dipakai untuk read yang butuh data relasi dengan tabel users.
// Bukan representasi tabel, murni untuk response.
type ReviewDetail struct {
	ID        uuid.UUID  `json:"review_id"`
	ProductID uuid.UUID  `json:"review_product_id"`
	Rating    int16      `json:"review_rating"`
	Comment   *string    `json:"review_comment"`
	CreatedAt time.Time  `json:"review_created_at"`
	User      ReviewUser `json:"user"`
}

type UpdateReviewRequest struct {
	Rating  *int16  `json:"rating" validate:"omitempty,min=1,max=5"`
	Comment *string `json:"comment" validate:"omitempty,max=255"`
}

type CreateReviewRequest struct {
	ProductId string  `json:"product_id" validate:"required,uuid"`
	Rating    int16   `json:"rating" validate:"required,min=1,max=5"`
	Comment   *string `json:"comment" validate:"omitempty"`
}
