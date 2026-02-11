package products

import (
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/pkg"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const maxFileSize = 4 * 1024 * 1024

type ProductsHandler struct {
	service   *ProductsService
	decoder   *form.Decoder
	validator *validator.Validate
}

func NewProductsHandler(svc *ProductsService, dec *form.Decoder, vld *validator.Validate) *ProductsHandler {
	return &ProductsHandler{
		service:   svc,
		decoder:   dec,
		validator: vld,
	}
}

func (ph *ProductsHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ph.service.GetAllProducts(r.Context())
	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (ph *ProductsHandler) AddNewProductsHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		pkg.JSONError(w, 400, "gagal parse form data")
		return
	}

	defer r.MultipartForm.RemoveAll()

	var productReq dto.CreateProductRequest

	if err := ph.decoder.Decode(&productReq, r.MultipartForm.Value); err != nil {
		fmt.Println(err)
		pkg.JSONError(w, 400, "form data tidak valid")
		return
	}

	productReq.ProductImages = r.MultipartForm.File["product_images"]
	for _, fileHeader := range productReq.ProductImages {
		if fileHeader.Size > maxFileSize {
			pkg.JSONError(w, 400, "gambar terlalu besar! maksimal 4mb")
			return
		}
	}

	if len(productReq.ProductImages) > 7 || len(productReq.ProductImages) < 1 {
		pkg.JSONError(w, 400, "minmal 1 dan maksimal 7 gambar")
		return
	}

	if err := ph.validator.Struct(productReq); err != nil {
		pkg.JSONError(w, 400, pkg.ValidationErrorsToMap(err))
		return
	}

	data, savedImages, insertErr := ph.service.Create(r.Context(), productReq)

	if insertErr != nil {
		pkg.JSONError(w, insertErr.Code, insertErr.Message)
		return
	}

	rsp := map[string]any{
		"products_data":   data,
		"products_images": savedImages,
	}

	pkg.JSONSuccess(w, 200, "berhasil menambahkan data!", rsp)

}

func (ph *ProductsHandler) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {

	idParam := chi.URLParam(r, "id")
	productId, err := uuid.Parse(idParam)
	if err != nil {
		pkg.JSONError(w, 400, "ID tidak valid")
		return
	}

	delErr := ph.service.DeleteProducts(r.Context(), productId)

	if delErr != nil {
		pkg.JSONError(w, delErr.Code, delErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menghapus data", nil)

}

func (ph *ProductsHandler) GetByIdHandler(w http.ResponseWriter, r *http.Request) {

	param := chi.URLParam(r, "id")
	productId, err := uuid.Parse(param)
	if err != nil {
		pkg.JSONError(w, 400, "Parameter tidak valid!")
		return
	}

	data, searchErr := ph.service.GetById(r.Context(), productId)

	if searchErr != nil {
		pkg.JSONError(w, searchErr.Code, searchErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)

}

func (ph *ProductsHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	if slug == "" {
		pkg.JSONError(w, 400, "slug tidak valid!")
		return
	}

	data, err := ph.service.GetBySlug(r.Context(), slug)

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (ph *ProductsHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {

	categoryName := chi.URLParam(r, "cat")

	if categoryName == "" {
		pkg.JSONError(w, 400, "nama kategory tidak valid!")
		return
	}

	data, err := ph.service.GetByCategorySlug(r.Context(), categoryName)

	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)
}

func (ph *ProductsHandler) SetUpRoute(r chi.Router) {

	r.Route("/products", func(r chi.Router) {

		r.Get("/get-all", ph.GetAllHandler)
		r.Get("/id/{id}", ph.GetByIdHandler)
		r.Get("/slug/{slug}", ph.GetBySlug)

		r.Get("/category/{cat}", ph.GetByCategory)

		r.Group(func(r chi.Router) {
			r.Post("/add", ph.AddNewProductsHandler)
			r.Delete("/delete/{id}", ph.DeleteProductHandler)

		})

	})
}
