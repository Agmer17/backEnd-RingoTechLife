package configs

import (
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/order"
	"backEnd-RingoTechLife/internal/productimage"
	"backEnd-RingoTechLife/internal/products"
	"backEnd-RingoTechLife/internal/review"
	"backEnd-RingoTechLife/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryConfigs struct {
	UserRepository         *user.UserRepositoryImpl
	CategoryRepository     *category.CategoryRepositoryImpl
	ProductsRepository     *products.ProductRepositoryImpl
	ProductImageRepository *productimage.ProductImageRepoImpl
	ReviewRepository       *review.ReviewRepositoryImpl
	OrderRepository        *order.OrderRepositoryImpl
}

func NewRepositoryConfigs(pool *pgxpool.Pool) *RepositoryConfigs {

	userRepo := user.NewUserRepository(pool)
	categoryRepo := category.NewCategoryRepository(pool)
	productRepo := products.NewProductsRepository(pool)
	productImgRepo := productimage.NewProductImageRepository(pool)
	reviewRepo := review.NewReviewRepository(pool)
	orderRepo := order.NewOrderRepository(pool)

	return &RepositoryConfigs{
		UserRepository:         userRepo,
		CategoryRepository:     categoryRepo,
		ProductsRepository:     productRepo,
		ProductImageRepository: productImgRepo,
		ReviewRepository:       reviewRepo,
		OrderRepository:        orderRepo,
	}

}
