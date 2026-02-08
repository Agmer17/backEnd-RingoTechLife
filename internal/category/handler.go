package category

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

type CategoryHandler struct {
	service   *CategoryService
	validator *validator.Validate
}

func NewCategoryHandler(svc *CategoryService, vld *validator.Validate) *CategoryHandler {
	return &CategoryHandler{
		service:   svc,
		validator: vld,
	}
}

func (c *CategoryHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {

	data, err := c.service.GetAllCategories(r.Context())

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)

}

func (c *CategoryHandler) GetByIdHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	if idParam == "" {
		pkg.JSONError(w, 400, "id tidak ditemukan di parameter! kirim data dengan benar!")
		return
	}
	catId, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	data, getErr := c.service.GetById(r.Context(), catId)

	if getErr != nil {
		pkg.JSONError(w, getErr.Code, getErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (c *CategoryHandler) GetBySlugHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	if slug == "" {
		pkg.JSONError(w, 400, "slug tidak ditemukan di parameter! kirim data dengan benar!")
		return
	}

	data, err := c.service.GetBySlug(r.Context(), slug)

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengabil data", data)

}

func (c *CategoryHandler) AddNewCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var categoryRequest dto.CreateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		pkg.JSONError(w, 400, "Harap isi data dengan benar!")
		return
	}

	if err := c.validator.Struct(categoryRequest); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	categoryModel := model.Category{
		Name:        categoryRequest.Name,
		Slug:        categoryRequest.Slug,
		Description: categoryRequest.Desc,
	}
	newCategory, insertErr := c.service.CreateNewCategory(r.Context(), categoryModel)

	if insertErr != nil {
		pkg.JSONError(w, insertErr.Code, insertErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menambah data", newCategory)
}

func (c *CategoryHandler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {

	idParam := chi.URLParam(r, "id")

	if idParam == "" {
		pkg.JSONError(w, 400, "id tidak ditemukan di parameter! kirim data dengan benar!")
		return
	}
	catId, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	var req dto.UpdateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.JSONError(w, 400, "Harap isi data dengan benar!")
		return
	}

	if err := c.validator.Struct(req); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)

		pkg.JSONError(w, 400, validationErr)
		return
	}

	data, updateErr := c.service.UpdateCategories(r.Context(), catId, req)

	if updateErr != nil {
		pkg.JSONError(w, updateErr.Code, updateErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengupdate data", data)

}

func (c *CategoryHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {

	idParam := chi.URLParam(r, "id")

	if idParam == "" {
		pkg.JSONError(w, 400, "id tidak ditemukan di parameter! kirim data dengan benar!")
		return
	}
	catId, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	delErr := c.service.DeleteCategory(r.Context(), catId)

	if delErr != nil {
		pkg.JSONError(w, delErr.Code, delErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menghapus data category", nil)
}

func (c *CategoryHandler) SetUpRoute(router chi.Router) {

	router.Route("/categories", func(r chi.Router) {

		r.Get("/get-all", c.GetAllHandler)
		r.Get("/slug/{slug}", c.GetBySlugHandler)
		r.Get("/id/{id}", c.GetByIdHandler)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Use(middleware.RoleMiddleware("ADMIN"))

			r.Post("/add", c.AddNewCategoryHandler)
			r.Delete("/delete/{id}", c.DeleteCategoryHandler)
			r.Put("/update/{id}", c.UpdateCategoryHandler)
		})

	})

}
