package auth

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/user"
	"backEnd-RingoTechLife/pkg"
	"context"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserService *user.UserService
}

func NewAuthService(userSvc *user.UserService) *AuthService {

	return &AuthService{
		UserService: userSvc,
	}

}

// set cookie
func (a *AuthService) Login(ctx context.Context, req LoginRequest, w http.ResponseWriter) *common.ErrorResponse {

	exist, err := a.UserService.ExistByEmailOrPhone(ctx, req.EmailOrPhone, req.EmailOrPhone)

	if err != nil {
		return common.NewErrorResponse(500, "gagal mengambil data ke database : "+err.Error())
	}

	if !exist {
		return common.NewErrorResponse(400, "user tidak ditemukan!")
	}

	userData, getErr := a.UserService.GetByEmailOrPhone(ctx, req.EmailOrPhone, req.EmailOrPhone)

	if getErr != nil {
		return getErr
	}

	hashErr := bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(req.Password))

	if hashErr != nil {
		return common.NewErrorResponse(401, "password salah!")
	}

	token, err := pkg.GenerateToken(
		userData.ID,
		userData.Role,
		60,
	)
	if err != nil {
		return common.NewErrorResponse(500, "gagal generate token")
	}

	// 🍪 SET COOKIE
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // WAJIB true kalau https
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 36000,
	})

	return nil

}
