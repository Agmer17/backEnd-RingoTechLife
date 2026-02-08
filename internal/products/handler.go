package products

import (
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/pkg"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

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

	if err := ph.validator.Struct(productReq); err != nil {
		pkg.JSONError(w, 400, pkg.ValidationErrorsToMap(err))
		return
	}

	data, insertErr := ph.service.Create(r.Context(), productReq)

	if insertErr != nil {
		pkg.JSONError(w, insertErr.Code, insertErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil menambahkan data!", data)

}

func (ph *ProductsHandler) SetUpRoute(r chi.Router) {

	r.Route("/products", func(r chi.Router) {

		r.Get("/get-all", ph.GetAllHandler)

		r.Group(func(r chi.Router) {

			r.Post("/add", ph.AddNewProductsHandler)

		})

	})
}
