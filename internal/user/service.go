package user

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const userFilePlace = "user"

type UserService struct {
	userRepo    UserRepositoryInterface
	FileStorage *storage.FileStorage
}

func NewUserService(userRepo UserRepositoryInterface, fileStorage *storage.FileStorage) *UserService {
	return &UserService{
		userRepo:    userRepo,
		FileStorage: fileStorage,
	}
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest) (model.User, *common.ErrorResponse) {

	exist, err := s.userRepo.IsUserExistsByEmailOrPhone(ctx, req.Email, req.PhoneNumber, nil)

	if exist {
		return model.User{}, common.NewErrorResponse(409, "Nomor telefon atau email telah terdaftar")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)

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
	return s.userRepo.IsUserExistsByEmailOrPhone(ctx, email, phone, nil)
}

func (s *UserService) Update(
	ctx context.Context,
	req dto.UpdateUserRequest,
	id uuid.UUID,
) (model.User, *common.ErrorResponse) {

	exist, userData, err := s.userRepo.IsUserExistsById(ctx, id)
	if err != nil {
		return model.User{}, common.NewErrorResponse(500, "gagal mengambil data dari database! "+err.Error())
	}

	if !exist {
		return model.User{}, common.NewErrorResponse(404, "user tidak ditemukan!")
	}

	// ================= UPDATE FIELDS =================
	if req.FullName != nil {
		userData.FullName = *req.FullName
	}

	if req.Email != nil {
		exist, err := s.userRepo.IsUserExistsByEmailOrPhone(ctx, *req.Email, "", &userData.ID)
		if err != nil {
			return model.User{}, common.NewErrorResponse(500, "gagal mengambil data dari database!"+err.Error())
		}

		if exist {
			return model.User{}, common.NewErrorResponse(400, "email sudah digunakan!")
		}

		userData.Email = *req.Email
	}

	if req.PhoneNumber != nil {
		exist, err := s.userRepo.IsUserExistsByEmailOrPhone(ctx, "", *req.PhoneNumber, &userData.ID)
		if err != nil {
			return model.User{}, common.NewErrorResponse(500, err.Error())
		}

		if exist {
			return model.User{}, common.NewErrorResponse(400, "nomor hp  "+*req.PhoneNumber+" sudah digunakan")
		}

		userData.PhoneNumber = req.PhoneNumber
	}

	if req.Role != nil {
		userData.Role = *req.Role
	}

	// nanti dari handler bakal ngisi datanya!
	//dibawah tinggal handle kalo gagal diapus aja filenya!
	var oldProfilePic *string
	if userData.ProfilePicture != nil {
		tmp := *userData.ProfilePicture
		oldProfilePic = &tmp
	}

	if req.DeleteProfilePicture != nil && *req.DeleteProfilePicture {
		userData.ProfilePicture = nil
	}

	if req.ProfilePicture != nil {
		userData.ProfilePicture = req.ProfilePicture
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), 10)
		if err != nil {
			return model.User{}, common.NewErrorResponse(500, "Internal server error! gagal saat menyimpan user : "+err.Error())
		}

		userData.Password = string(hashedPassword)
	}

	// ================= SAVE =================
	updatedUser, err := s.userRepo.Update(ctx, &userData)
	if err != nil {
		if req.ProfilePicture != nil {
			s.FileStorage.DeletePublicFile(*req.ProfilePicture, userFilePlace)
		}
		return model.User{}, common.NewErrorResponse(500, "gagal update user! "+err.Error())
	}

	// kalo updatenya successfull apus file gambar lama!
	if oldProfilePic != nil &&
		(req.ProfilePicture != nil ||
			(req.DeleteProfilePicture != nil && *req.DeleteProfilePicture)) {

		s.FileStorage.DeletePublicFile(*oldProfilePic, userFilePlace)
	}

	return *updatedUser, nil
}

func (s *UserService) Delete(ctx context.Context, req dto.DeleteUserRequest) *common.ErrorResponse {
	user, getErr := s.GetByID(ctx, req.ID)
	if getErr != nil {
		return getErr
	}

	// delete image kalo ada!
	if user.ProfilePicture != nil && *user.ProfilePicture != "" {
		s.FileStorage.DeletePublicFile(userFilePlace, userFilePlace)
	}

	err := s.userRepo.Delete(ctx, req.ID)

	if err != nil {
		return common.NewErrorResponse(500, "something wrong with the database!"+err.Error())
	}

	return nil
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

func (s *UserService) GetAllUser(ctx context.Context) ([]dto.UserDataResponse, *common.ErrorResponse) {
	data, err := s.userRepo.GetAllUsers(ctx)

	if err != nil {
		return []dto.UserDataResponse{}, common.NewErrorResponse(500, "soemthing wrong with database"+err.Error())
	}

	var respData []dto.UserDataResponse = make([]dto.UserDataResponse, len(data))

	for i, v := range data {
		respData[i] = dto.ModelUserToResponse(v)
	}

	return respData, nil
}
