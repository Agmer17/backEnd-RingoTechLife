package review

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ReviewService struct {
	reviewRepo ReviewRepository
}

func NewReviewService(revRepo *ReviewRepositoryImpl) *ReviewService {
	return &ReviewService{
		reviewRepo: revRepo,
	}
}

func (r *ReviewService) Create(ctx context.Context, req model.Review) (dto.ReviewDetail, *common.ErrorResponse) {

	data, err := r.reviewRepo.Create(ctx, &req)

	if err != nil {

		if errors.Is(err, ErrReviewAlreadyExists) {
			return dto.ReviewDetail{}, common.NewErrorResponse(409, "kamu sudah pernah review produk ini!")
		}

		if errors.Is(err, ErrReviewUserNotFound) {
			return dto.ReviewDetail{}, common.NewErrorResponse(404, "user tidak ditemukan!")
		}

		if errors.Is(err, ErrReviewProductNotFound) {
			return dto.ReviewDetail{}, common.NewErrorResponse(404, "produk yang di review tidak ditemukan! mungkin sudah dihapus!")
		}
	}

	reviewResp, err := r.reviewRepo.GetDetailByID(ctx, data.ID)
	if err != nil {

		if errors.Is(err, ErrReviewNotFound) {
			return dto.ReviewDetail{}, common.NewErrorResponse(404, "Terjadi kesalahan di databse! review tidak ditemukan")
		}

		return dto.ReviewDetail{}, common.NewErrorResponse(500, "internal server error!")
	}

	return *reviewResp, nil
}

func (r *ReviewService) GetWithDetailProdId(ctx context.Context, productId uuid.UUID) ([]dto.ReviewDetail, *common.ErrorResponse) {

	data, err := r.reviewRepo.GetAllDetailsProdId(ctx, productId)
	if err != nil {

		return []dto.ReviewDetail{}, common.NewErrorResponse(500, "internal server error! gagal mengambil data ke database!")
	}

	return data, nil

}

func (r *ReviewService) Delete(ctx context.Context, reviewId uuid.UUID) *common.ErrorResponse {

	err := r.reviewRepo.Delete(ctx, reviewId)

	if err != nil {
		if errors.Is(err, ErrReviewNotFound) {
			return common.NewErrorResponse(404, "review tidak ditemukan!")
		}

		return common.NewErrorResponse(500, "internal server error! gagal mengambil data ke database")
	}

	return nil
}

func (r *ReviewService) Update(ctx context.Context, reviewId uuid.UUID, userId uuid.UUID, updateReq dto.UpdateReviewRequest) (dto.ReviewDetail, *common.ErrorResponse) {

	// get detail data lama dulu!
	data, err := r.reviewRepo.GetDetailByID(ctx, reviewId)
	if err != nil {
		if errors.Is(err, ErrReviewNotFound) {
			return dto.ReviewDetail{}, common.NewErrorResponse(404, "review tidak ditemukan!")
		}
		return dto.ReviewDetail{}, common.NewErrorResponse(500, "internal server error!")
	}

	if userId != data.User.ID {
		return dto.ReviewDetail{}, common.NewErrorResponse(401, "kamu tidak bisa mengedit review ini!")
	}

	updatedModel := model.Review{
		ID:        data.ID,
		ProductID: data.ProductID,
		UserID:    data.User.ID,
		Rating:    data.Rating,
		Comment:   data.Comment,
		CreatedAt: data.CreatedAt,
	}

	if updateReq.Comment != nil && *updateReq.Comment != "" {
		updatedModel.Comment = updateReq.Comment
		data.Comment = updateReq.Comment
	}

	if updateReq.Rating != nil && *updateReq.Rating > 0 {
		updatedModel.Rating = *updateReq.Rating
		data.Rating = *updateReq.Rating
	}

	_, err = r.reviewRepo.Update(ctx, &updatedModel)

	if err != nil {

		if errors.Is(err, ErrReviewNotFound) {
			return dto.ReviewDetail{}, common.NewErrorResponse(404, "data review tidak ditemukan! mungkin sudah dihapus!")
		}

		return dto.ReviewDetail{}, common.NewErrorResponse(500, "internal server error!")
	}

	return *data, nil
}
