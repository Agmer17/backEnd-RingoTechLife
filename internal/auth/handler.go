package auth

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	EmailOrPhone string `json:"email_or_phone"`
	Password     string `json:"password"`
}

type AuthHandler struct {
	AuthService *AuthService
	Validator   *validator.Validate
}

func NewAuthHandler(svc *AuthService, validator *validator.Validate) *AuthHandler {
	return &AuthHandler{
		AuthService: svc,
		Validator:   validator,
	}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {

	reqCtx := r.Context()

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Body tidak valid! harap masukan data dengan benar")
		return
	}

	data, err := h.AuthService.Login(reqCtx, req, w)

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, data.Message, data.Data)

}

func (h *AuthHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {

	var req dto.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Harap isi data dengan benar!")
		return
	}

	if err := h.Validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	successRes, err := h.AuthService.SignUp(r.Context(), req)

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, successRes.Message, successRes.Data)

}

func (h *AuthHandler) SetUpRoute(router chi.Router) {

	router.Route("/auth", func(r chi.Router) {
		r.Use(httprate.Limit(
			20,
			time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				errorRes := common.NewErrorResponse(http.StatusTooManyRequests, "Tolong jangan lakukan request terlalu banyak")

				errorResJson, _ := json.Marshal(errorRes)

				w.Write(errorResJson)

			}),
		))

		r.Post("/login", h.LoginHandler)
		r.Post("/sign-up", h.SignupHandler)
	})
}
