package configs

import (
	"backEnd-RingoTechLife/internal/auth"
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/common"
	"backEnd-RingoTechLife/internal/order"
	"backEnd-RingoTechLife/internal/products"
	"backEnd-RingoTechLife/internal/review"
	"backEnd-RingoTechLife/internal/user"
	"backEnd-RingoTechLife/pkg"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

func SetupRouter(r chi.Router, svcCfg *ServiceConfigs) {

	// ======= validator
	validator := validator.New()
	validator.RegisterValidation("phoneID", pkg.PhoneID)
	validator.RegisterValidation("slug", pkg.SlugValidator)

	// ==== form decoder
	decoder := form.NewDecoder()

	authHandler := auth.NewAuthHandler(svcCfg.AuthService, validator)
	userHandler := user.NewUserHandler(svcCfg.UserService, decoder, validator)
	categoryHandler := category.NewCategoryHandler(svcCfg.CategoryService, validator)
	productHandler := products.NewProductsHandler(svcCfg.ProductService, decoder, validator)
	reviewHandler := review.NewReviewHandler(svcCfg.ReviewService, validator)
	orderHandler := order.NewOrderHandler(svcCfg.OrderService, validator)

	fileServer := http.FileServer(http.Dir(svcCfg.ServerStorage.Public))

	r.Use(httprate.Limit(
		200,
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

	r.Route("/api", func(r chi.Router) {
		authHandler.SetUpRoute(r)
		userHandler.SetUpRoute(r)
		categoryHandler.SetUpRoute(r)
		productHandler.SetUpRoute(r)
		reviewHandler.SetupRoute(r)
		orderHandler.SetUpRoute(r)
	})

	r.Handle("/uploads/public/*", http.StripPrefix("/uploads/public/", fileServer))
}
