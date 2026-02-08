package user

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/middleware"
	"backEnd-RingoTechLife/internal/storage"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	maxFileSize = 5 << 20 // 5MB
)

type UserHandler struct {
	UserService *UserService
	FileStorage *storage.FileStorage
	Validator   *validator.Validate
}

func NewUserHandler(svc *UserService, fileStorage *storage.FileStorage, validator *validator.Validate) *UserHandler {
	return &UserHandler{
		UserService: svc,
		FileStorage: fileStorage,
		Validator:   validator,
	}
}

// ==================== USER ENDPOINTS ====================

// GetCurrentUserHandler - GET /user/profile
func (h *UserHandler) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		pkg.JSONError(w, 401, "User ID tidak ditemukan")
		return
	}

	user, err := h.UserService.GetByID(r.Context(), userID)
	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	responseData := dto.ModelUserToResponse(user)

	pkg.JSONSuccess(w, 200, "Berhasil mengambil data user", responseData)
}

// UpdateCurrentUserHandler - PUT /user/profile
func (h *UserHandler) UpdateCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		pkg.JSONError(w, 401, "User ID tidak ditemukan")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		pkg.JSONError(w, 400, "Ukuran file terlalu besar atau format tidak valid")
		return
	}

	var req dto.UpdateUserRequest
	req.ID = userID

	// Parse text fields
	if fullName := r.FormValue("full_name"); fullName != "" {
		if len(fullName) < 3 {
			pkg.JSONError(w, 400, "Full name minimal 3 karakter")
			return
		}
		req.FullName = &fullName
	}

	if email := r.FormValue("email"); email != "" {
		if err := h.Validator.Var(email, "email"); err != nil {
			pkg.JSONError(w, 400, "Format email tidak valid")
			return
		}
		req.Email = &email
	}

	if phoneNumber := r.FormValue("phone_number"); phoneNumber != "" {
		if err := h.Validator.Var(phoneNumber, "phoneID"); err != nil {
			pkg.JSONError(w, 400, "Format nomor telepon tidak valid")
			return
		}
		req.PhoneNumber = &phoneNumber
	}

	if password := r.FormValue("password"); password != "" {
		if len(password) < 8 {
			pkg.JSONError(w, 400, "Password minimal 8 karakter")
			return
		}
		req.Password = &password
	}

	if role := r.FormValue("role"); role != "" {
		pkg.JSONError(w, 401, "UNAUTHORIZED!")
		return
	}

	var newFileName string
	// var oldProfilePicture *string

	// Handle file upload
	file, fileHeader, err := r.FormFile("profile_picture")
	if err == nil {
		defer file.Close()

		// Validate file size
		if fileHeader.Size > maxFileSize {
			pkg.JSONError(w, 413, "Ukuran file maksimal 5MB")
			return
		}

		// Detect and validate mime type
		mimeType, err := h.FileStorage.DetectFileType(fileHeader)
		if err != nil {
			pkg.JSONError(w, 500, "Gagal membaca file")
			return
		}

		fileExt, supported := h.FileStorage.IsTypeSupportted(mimeType)
		if !supported {
			pkg.JSONError(w, 400, "Format file tidak didukung. Gunakan JPEG, PNG, atau WEBP")
			return
		}

		mediaType := h.FileStorage.GetMediaType(mimeType)
		if mediaType != "IMAGE" {
			pkg.JSONError(w, 400, "File harus berupa gambar")
			return
		}

		// Save new file
		fileName, err := h.FileStorage.SavePublicFile(fileHeader, fileExt, userFilePlace)
		if err != nil {
			pkg.JSONError(w, 500, "Gagal menyimpan file: "+err.Error())
			return
		}

		newFileName = fileName
		req.ProfilePicture = &fileName
	} else if err != http.ErrMissingFile {
		pkg.JSONError(w, 400, "Error saat membaca file")
		return
	}

	// Update user
	updatedUser, errUpdate := h.UserService.Update(r.Context(), req, userID)
	if errUpdate != nil {
		// Rollback: delete new file if service failed
		if newFileName != "" {
			h.FileStorage.DeletePublicFile(newFileName, userFilePlace)
		}
		pkg.JSONError(w, errUpdate.Code, errUpdate.Message)
		return
	}

	// Remove password from response
	updatedUser.Password = ""

	pkg.JSONSuccess(w, 200, "Berhasil update profile", updatedUser)
}

// DeleteCurrentUserHandler - DELETE /user/profile
func (h *UserHandler) DeleteCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		pkg.JSONError(w, 401, "User ID tidak ditemukan")
		return
	}

	// // Get user data first to get profile picture
	// user, errGet := h.UserService.GetByID(r.Context(), userID)
	// if errGet != nil {
	// 	pkg.JSONError(w, errGet.Code, errGet.Message)
	// 	return
	// }

	// Delete user from database
	err := h.UserService.Delete(r.Context(), dto.DeleteUserRequest{ID: userID})
	if err != nil {
		pkg.JSONError(w, 500, "Gagal menghapus user: "+err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menghapus user!", nil)

}

// ==================== ADMIN ENDPOINTS ====================

// GetUserByIDHandler - GET /user/{id}
func (h *UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	user, errGet := h.UserService.GetByID(r.Context(), userID)
	if errGet != nil {
		pkg.JSONError(w, errGet.Code, errGet.Message)
		return
	}

	// Remove password from response
	user.Password = ""

	pkg.JSONSuccess(w, 200, "Berhasil mengambil data user", user)
}

// UpdateUserByIDHandler - PUT /user/{id}
func (h *UserHandler) UpdateUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		pkg.JSONError(w, 400, "Ukuran file terlalu besar atau format tidak valid")
		return
	}

	var req dto.UpdateUserRequest
	req.ID = userID

	// Parse text fields
	if fullName := r.FormValue("full_name"); fullName != "" {
		if len(fullName) < 3 {
			pkg.JSONError(w, 400, "Full name minimal 3 karakter")
			return
		}
		req.FullName = &fullName
	}

	if email := r.FormValue("email"); email != "" {
		if err := h.Validator.Var(email, "email"); err != nil {
			pkg.JSONError(w, 400, "Format email tidak valid")
			return
		}
		req.Email = &email
	}

	if phoneNumber := r.FormValue("phone_number"); phoneNumber != "" {
		if err := h.Validator.Var(phoneNumber, "phoneID"); err != nil {
			pkg.JSONError(w, 400, "Format nomor telepon tidak valid")
			return
		}
		req.PhoneNumber = &phoneNumber
	}

	if password := r.FormValue("password"); password != "" {
		if len(password) < 8 {
			pkg.JSONError(w, 400, "Password minimal 8 karakter")
			return
		}
		req.Password = &password
	}

	if role := r.FormValue("role"); role != "" {
		if role != "ADMIN" && role != "USER" {
			pkg.JSONError(w, 400, "Role harus ADMIN atau USER")
			return
		}
		req.Role = &role
	}

	if deleteImage := r.FormValue("delete_profile_picture"); deleteImage != "" {

		parsed, err := strconv.ParseBool(deleteImage)

		if err != nil {
			pkg.JSONError(w, 400, "deleted profile picture harus boolean!")
			return
		}

		req.DeleteProfilePicture = &parsed

	}

	var newFileName string

	// Handle file upload
	file, fileHeader, err := r.FormFile("profile_picture")
	if err == nil {
		defer file.Close()

		// Validate file size
		if fileHeader.Size > maxFileSize {
			pkg.JSONError(w, 413, "Ukuran file maksimal 5MB")
			return
		}

		// Detect and validate mime type
		mimeType, err := h.FileStorage.DetectFileType(fileHeader)
		if err != nil {
			pkg.JSONError(w, 500, "Gagal membaca file")
			return
		}

		fileExt, supported := h.FileStorage.IsTypeSupportted(mimeType)
		if !supported {
			pkg.JSONError(w, 400, "Format file tidak didukung. Gunakan JPEG, PNG, atau WEBP")
			return
		}

		mediaType := h.FileStorage.GetMediaType(mimeType)
		if mediaType != "IMAGE" {
			pkg.JSONError(w, 400, "File harus berupa gambar")
			return
		}

		// Save new file
		fileName, err := h.FileStorage.SavePublicFile(fileHeader, fileExt, userFilePlace)
		if err != nil {
			pkg.JSONError(w, 500, "Gagal menyimpan file: "+err.Error())
			return
		}

		newFileName = fileName
		req.ProfilePicture = &fileName
	} else if err != http.ErrMissingFile {
		pkg.JSONError(w, 400, "Error saat membaca file")
		return
	}

	// Update user
	updatedUser, errUpdate := h.UserService.Update(r.Context(), req, userID)
	if errUpdate != nil {
		// Rollback: delete new file if service failed
		if newFileName != "" {
			h.FileStorage.DeletePublicFile(newFileName, userFilePlace)
		}
		pkg.JSONError(w, errUpdate.Code, errUpdate.Message)
		return
	}

	// Remove password from response
	respUserData := dto.ModelUserToResponse(updatedUser)

	pkg.JSONSuccess(w, 200, "Berhasil update user", respUserData)
}

// DeleteUserByIDHandler - DELETE /user/{id}
func (h *UserHandler) DeleteUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	// Delete user from database
	errDelete := h.UserService.Delete(r.Context(), dto.DeleteUserRequest{ID: userID})
	if errDelete != nil {
		pkg.JSONError(w, errDelete.Code, errDelete.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "Berhasil menghapus user", nil)
}

func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {

	data, err := h.UserService.GetAllUser(r.Context())
	if err != nil {
		pkg.JSONError(w, 500, "internal server error"+err.Message)
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (h *UserHandler) AddNewUserHandler(w http.ResponseWriter, r *http.Request) {

	var req dto.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "invalid request body! harap isi data dengan benar!")
		return
	}

	if err := h.Validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	data, insertErr := h.UserService.Create(r.Context(), req)

	if insertErr != nil {
		pkg.JSONError(w, insertErr.Code, insertErr.Message)
		return
	}

	respData := dto.ModelUserToResponse(data)

	pkg.JSONSuccess(w, 200, "berhasil menambahkan user", respData)
}

// ==================== SETUP ROUTES ====================

func (h *UserHandler) SetUpRoute(router chi.Router) {
	router.Route("/user", func(r chi.Router) {
		// Rate limiting
		r.Use(httprate.Limit(
			30,
			time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				errorRes := common.NewErrorResponse(http.StatusTooManyRequests, "Terlalu banyak request, coba lagi nanti")
				errorResJson, _ := json.Marshal(errorRes)
				w.Write(errorResJson)
			}),
		))

		// User endpoints (authenticated users only)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			r.Get("/profile/me", h.GetCurrentUserHandler)
			r.Put("/profile", h.UpdateCurrentUserHandler)
			r.Delete("/profile", h.DeleteCurrentUserHandler)
		})

		// Admin endpoints (admin only)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Use(middleware.RoleMiddleware("ADMIN"))

			r.Get("/id/{id}", h.GetUserByIDHandler)
			r.Put("/{id}", h.UpdateUserByIDHandler)
			r.Get("/get-all", h.GetAllUsersHandler)
			r.Delete("/{id}", h.DeleteUserByIDHandler)
			r.Post("/add", h.AddNewUserHandler)
		})
	})
}
