package configs

import (
	"backEnd-RingoTechLife/internal/common"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type RoutesHandler struct {
}

func NewRouter(r *chi.Mux) *RoutesHandler {

	r.Use(httprate.Limit(
		150,
		time.Minute,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			errorRes := common.NewErrorResponse(http.StatusTooManyRequests, "Tolong jangan lakukan request terlalu banyak")

			errorResJson, _ := json.Marshal(errorRes)

			w.Write(errorResJson)

		}),
	))

	r.Use(middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  log.Default(),
			NoColor: false}))

	r.Use(middleware.Recoverer)
	return &RoutesHandler{}
}
