package configs

import (
	"backEnd-RingoTechLife/internal/auth"
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/order"
	"backEnd-RingoTechLife/internal/payment"
	"backEnd-RingoTechLife/internal/productimage"
	"backEnd-RingoTechLife/internal/products"
	"backEnd-RingoTechLife/internal/review"
	"backEnd-RingoTechLife/internal/storage"
	"backEnd-RingoTechLife/internal/user"
	"context"
)

type ServiceConfigs struct {
	AuthService     *auth.AuthService
	UserService     *user.UserService
	ServerStorage   *storage.FileStorage
	CategoryService *category.CategoryService
	ProductService  *products.ProductsService
	ReviewService   *review.ReviewService
	OrderService    *order.OrderService
	PaymentService  *payment.PayementService
}

func NewServiceConfigs(rcf *RepositoryConfigs, serverStorage *storage.FileStorage) *ServiceConfigs {

	serviceContext := context.Background()

	userSvc := user.NewUserService(rcf.UserRepository, serverStorage)
	authSvc := auth.NewAuthService(userSvc)
	categorySvc := category.NewCategoryService(rcf.CategoryRepository)
	productImageSvc := productimage.NewProductImageService(rcf.ProductImageRepository, serverStorage)
	productSvc := products.NewProductsService(rcf.ProductsRepository, serverStorage, productImageSvc)
	reviewsSvc := review.NewReviewService(rcf.ReviewRepository)
	orderSvc := order.NewOrderService(rcf.OrderRepository, productSvc, serviceContext)
	paymentSvc := payment.NewPaymentService(rcf.PaymentRepository, serverStorage)

	return &ServiceConfigs{
		AuthService:     authSvc,
		UserService:     userSvc,
		ServerStorage:   serverStorage,
		CategoryService: categorySvc,
		ProductService:  productSvc,
		ReviewService:   reviewsSvc,
		OrderService:    orderSvc,
		PaymentService:  paymentSvc,
	}

}
