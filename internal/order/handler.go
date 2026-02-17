package order

import (
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/middleware"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type OrderHandler struct {
	transactionService *TransactionService
}

func NewOrderHandler(tsvc *TransactionService) *OrderHandler {

	return &OrderHandler{
		transactionService: tsvc,
	}
}

func (t *OrderHandler) SetUpRoute(router chi.Router) {

	router.Route("/transaction", func(r chi.Router) {
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
		r.Use(middleware.AuthMiddleware)
	})

}
