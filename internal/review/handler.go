package review

import (
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/common/model"
	"backEnd-RingoTechLife/internal/middleware"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	reviewService *ReviewService
	validator     *validator.Validate
}

func NewReviewHandler(svc *ReviewService, vld *validator.Validate) *ReviewHandler {
	return &ReviewHandler{
		reviewService: svc,
		validator:     vld,
	}
}

func (rh *ReviewHandler) HandlerCreate(w http.ResponseWriter, r *http.Request) {

	var addReq dto.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&addReq); err != nil {
		pkg.JSONError(w, 400, "data yang kamu kirimkan tidak valid! harap isi semuanya!")
		return
	}

	if err := rh.validator.Struct(addReq); err != nil {
		pkg.JSONError(w, 400, pkg.ValidationErrorsToMap(err))
		return
	}

	prodId, err := uuid.Parse(addReq.ProductId)
	if err != nil {
		pkg.JSONError(w, 400, "gagal parsing uuid! "+err.Error())
		return
	}

	userId, _ := middleware.GetUserID(r.Context())

	tempModel := model.Review{
		ProductID: prodId,
		UserID:    userId,
		Rating:    addReq.Rating,
		Comment:   addReq.Comment,
	}

	result, addErr := rh.reviewService.Create(r.Context(), tempModel)

	if addErr != nil {
		pkg.JSONError(w, addErr.Code, addErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menambahkan review", result)
}

func (rh *ReviewHandler) GetReviewFromProduct(w http.ResponseWriter, r *http.Request) {

	productId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		pkg.JSONError(w, 400, "id produk tidak valid!")
		return
	}

	data, getErr := rh.reviewService.GetWithDetailProdId(r.Context(), productId)
	if getErr != nil {
		pkg.JSONError(w, getErr.Code, getErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data review!", data)

}

func (rh *ReviewHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {

	deletedProductId, err := uuid.Parse(chi.URLParam(r, "reviewId"))
	if err != nil {
		pkg.JSONError(w, 400, "id review tidak valid!")
		return
	}

	delErr := rh.reviewService.Delete(r.Context(), deletedProductId)

	if delErr != nil {
		pkg.JSONError(w, delErr.Code, delErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menghapus data", nil)

}

func (rh *ReviewHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	var updateReq dto.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		pkg.JSONError(w, 400, "data yang kamu kirim tidak valid!")
		return
	}

	if err := rh.validator.Struct(updateReq); err != nil {
		pkg.JSONError(w, 400, pkg.ValidationErrorsToMap(err))
		return
	}

	prodId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		pkg.JSONError(w, 400, "id tidak valid!")
		return
	}

	userId, _ := middleware.GetUserID(r.Context())

	result, updateErr := rh.reviewService.Update(r.Context(), prodId, userId, updateReq)

	if updateErr != nil {
		pkg.JSONError(w, updateErr.Code, updateErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengupdate data", result)

}

func (rh *ReviewHandler) GetAllUserReviewHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := middleware.GetUserID(r.Context())
	if !ok {
		pkg.JSONError(w, 401, "Harap logi terlebih dahulu sebelum mengakses fitur ini")
		return
	}

	data, err := rh.reviewService.GetReviewFromUser(r.Context(), userId)
	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)

}

func (rh *ReviewHandler) DeleteMyReviewHandler(w http.ResponseWriter, r *http.Request) {

	deletedProductId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		pkg.JSONError(w, 400, "id review tidak valid!")
		return
	}

	userId, ok := middleware.GetUserID(r.Context())

	if !ok {
		pkg.JSONError(w, 401, "sesi kamu sudah habis!")
		return
	}

	delErr := rh.reviewService.DeleteCurrentUserReview(r.Context(), deletedProductId, userId)

	if delErr != nil {
		pkg.JSONError(w, delErr.Code, delErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menghapus data", nil)

}

func (rh *ReviewHandler) getAllReviewHandler(w http.ResponseWriter, r *http.Request) {

	data, err := rh.reviewService.GetAllReview(r.Context())

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (rh *ReviewHandler) SetupRoute(router chi.Router) {

	router.Route("/reviews", func(r chi.Router) {

		r.Get("/product/{id}", rh.GetReviewFromProduct)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			r.Post("/create", rh.HandlerCreate)
			r.Put("/update/{id}", rh.UpdateHandler)
			r.Get("/get-my-review", rh.GetAllUserReviewHandler)
			r.Delete("/delete-my-review/{id}", rh.DeleteMyReviewHandler)

		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Use(middleware.RoleMiddleware(middleware.RoleAdmin))

			r.Delete("/delete/{reviewId}", rh.DeleteHandler)
			r.Get("/get-all", rh.getAllReviewHandler)
		})

	})
}
