package auth

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/user"
	"backEnd-RingoTechLife/pkg"
	"context"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const SevenDays = 7 * 24 * 60 * 60

type SignupResponse struct {
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthService struct {
	UserService *user.UserService
}

func NewAuthService(userSvc *user.UserService) *AuthService {

	return &AuthService{
		UserService: userSvc,
	}

}

// set cookie
func (a *AuthService) Login(ctx context.Context, req LoginRequest, w http.ResponseWriter) (common.SuccessResponse, *common.ErrorResponse) {

	exist, err := a.UserService.ExistByEmailOrPhone(ctx, req.EmailOrPhone, req.EmailOrPhone)

	if err != nil {
		return common.SuccessResponse{}, common.NewErrorResponse(500, "gagal mengambil data ke database : "+err.Error())
	}

	if !exist {
		return common.SuccessResponse{}, common.NewErrorResponse(400, "user tidak ditemukan!")
	}

	userData, getErr := a.UserService.GetByEmailOrPhone(ctx, req.EmailOrPhone, req.EmailOrPhone)

	if getErr != nil {
		return common.SuccessResponse{}, getErr
	}

	hashErr := bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(req.Password))

	if hashErr != nil {
		return common.SuccessResponse{}, common.NewErrorResponse(401, "password salah!")
	}

	refreshToken, err := pkg.GenerateTokenNoRole(
		userData.ID,
		168,
	)
	if err != nil {
		return common.SuccessResponse{}, common.NewErrorResponse(500, "gagal generate token")
	}

	token, err := pkg.GenerateToken(userData.ID, userData.Role, 60)

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   SevenDays,
	})

	return common.SuccessResponse{
		Message: "Berhasil login!",
		Data: map[string]string{
			"id":           userData.ID.String(),
			"role":         userData.Role,
			"access_token": token,
		},
	}, nil

}

func (a *AuthService) SignUp(ctx context.Context, req dto.CreateUserRequest) (common.SuccessResponse, *common.ErrorResponse) {

	data, err := a.UserService.Create(ctx, req)

	if err != nil {
		return common.SuccessResponse{}, err
	}

	successResponse := common.SuccessResponse{
		Message: "Berhasil membuat akun!",
		Data: SignupResponse{
			FullName:  data.FullName,
			CreatedAt: data.CreatedAt,
		},
	}

	return successResponse, nil

}
