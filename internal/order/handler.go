package order

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/common/dto"
	"backEnd-RingoTechLife/internal/middleware"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService *OrderService
	validator    *validator.Validate
}

func NewOrderHandler(osvc *OrderService, vld *validator.Validate) *OrderHandler {
	return &OrderHandler{
		orderService: osvc,
		validator:    vld,
	}
}

func (th *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {

	var order dto.CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		pkg.JSONError(w, 400, "harap isi data order dengan benar!")
		return
	}

	if err := th.validator.Struct(order); err != nil {
		validationErr := pkg.ValidationErrorsToMap(err)
		pkg.JSONError(w, 400, validationErr)
		return
	}

	produtId, err := uuid.Parse(order.ProductId)
	if err != nil {
		pkg.JSONError(w, 400, "product id tidak valid! gagal membuat order")
		return
	}

	userId, _ := middleware.GetUserID(r.Context())
	result, insertErr := th.orderService.CreateOneOrder(r.Context(), produtId, order.Quantity, userId)

	if insertErr != nil {
		pkg.JSONError(w, insertErr.Code, insertErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil membuat order", result)

}

func (th *OrderHandler) GetAllOfMyOrder(w http.ResponseWriter, r *http.Request) {

	userId, _ := middleware.GetUserID(r.Context())

	data, err := th.orderService.GetAllOrderByUserId(r.Context(), userId)
	if err != nil {
		pkg.JSONError(w, err.Code, err.Message)
		return
	}
	pkg.JSONSuccess(w, 200, "berhasil mengambil data", data)

}

func (th *OrderHandler) GetOrderById(w http.ResponseWriter, r *http.Request) {

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		pkg.JSONError(w, 400, "id tidak valid!")
		return
	}

	userId, _ := middleware.GetUserID(r.Context())

	data, getErr := th.orderService.GetByOrderId(r.Context(), id, userId)

	if getErr != nil {
		pkg.JSONError(w, getErr.Code, getErr.Message)
		return
	}

	pkg.JSONSuccess(w, 200, "berhasil mengambil data!", data)
}

func (th *OrderHandler) SetUpRoute(router chi.Router) {

	router.Route("/orders", func(r chi.Router) {
		r.Use(httprate.Limit(
			20,
			time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				errorRes := common.NewErrorResponse(http.StatusTooManyRequests, "Terlalu banyak request, coba lagi nanti")
				errorResJson, _ := json.Marshal(errorRes)
				w.Write(errorResJson)
			}),
		))
		r.Use(middleware.AuthMiddleware)

		r.Post("/create-order", th.CreateOrderHandler)
		r.Get("/my-orders", th.GetAllOfMyOrder)
		r.Get("/id/{id}", th.GetOrderById)

	})

}
