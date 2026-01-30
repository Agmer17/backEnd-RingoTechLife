package user

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo UserRepositoryInterface
}

func NewUserService(userRepo UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest) (model.User, *common.ErrorResponse) {

	exist, err := s.userRepo.IsUserExists(ctx, req.Email, req.PhoneNumber)

	if exist {
		return model.User{}, common.NewErrorResponse(409, "Nomor telefon atau email telah terdaftar")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)

	if err != nil {
		return model.User{}, common.NewErrorResponse(500, "Internal server error! gagal saat menyimpan user : "+err.Error())
	}

	user := &model.User{
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: &req.PhoneNumber,
		Password:    string(hashedPassword),
	}

	data, err := s.userRepo.Create(ctx, user)

	if err != nil {
		return *data, common.NewErrorResponse(500, "Gagal membuat akun! akun mungkin sudah ada! : "+err.Error())
	}

	return *data, nil
}

func (s *UserService) ExistByEmailOrPhone(ctx context.Context, email string, phone string) (bool, error) {
	return s.userRepo.IsUserExists(ctx, email, phone)
}

func (s *UserService) Update(
	ctx context.Context,
	req dto.UpdateUserRequest,
) (*model.User, *common.ErrorResponse) {

	exist, err := s.userRepo.IsUserExists(ctx, req.Email, req.PhoneNumber)

	if !exist {
		return &model.User{}, common.NewErrorResponse(404, "Nomor telefon atau email telah terdaftar")
	}

	user := &model.User{
		ID:             req.ID,
		FullName:       req.FullName,
		Email:          req.Email,
		PhoneNumber:    &req.PhoneNumber,
		Role:           req.Role,
		ProfilePicture: req.ProfilePicture,
	}

	data, err := s.userRepo.Update(ctx, user)

	if err != nil {
		return nil, common.NewErrorResponse(500, "gagal mengupdate data : "+err.Error())
	}

	return data, nil
}

func (s *UserService) Delete(ctx context.Context, req dto.DeleteUserRequest) error {
	return s.userRepo.Delete(ctx, req.ID)
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (model.User, *common.ErrorResponse) {

	data, err := s.userRepo.GetByID(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, common.NewErrorResponse(404, "Akun tidak dtemukan")
		}
		return model.User{}, common.NewErrorResponse(500, "Internal server error : "+err.Error())
	}

	return data, nil
}

func (s *UserService) GetByEmailOrPhone(ctx context.Context, email string, phoneNumber string) (model.User, *common.ErrorResponse) {
	data, err := s.userRepo.GetByEmailOrPhone(ctx, email, phoneNumber)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, common.NewErrorResponse(404, "Akun tidak dtemukan")
		}
		return model.User{}, common.NewErrorResponse(500, "Internal server error : "+err.Error())
	}

	return data, nil
}
