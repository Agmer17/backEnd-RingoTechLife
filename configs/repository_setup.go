package configs

import (
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/productimage"
	"backEnd-RingoTechLife/internal/products"
	"backEnd-RingoTechLife/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryConfigs struct {
	UserRepository         *user.UserRepositoryImpl
	CategoryRepository     *category.CategoryRepositoryImpl
	ProductsRepository     *products.ProductRepositoryImpl
	ProductImageRepository *productimage.ProductImageRepoImpl
}

func NewRepositoryConfigs(pool *pgxpool.Pool) *RepositoryConfigs {

	userRepo := user.NewUserRepository(pool)
	categoryRepo := category.NewCategoryRepository(pool)
	productRepo := products.NewProductsRepository(pool)
	productImgRepo := productimage.NewProductImageRepository(pool)

	return &RepositoryConfigs{
		UserRepository:         userRepo,
		CategoryRepository:     categoryRepo,
		ProductsRepository:     productRepo,
		ProductImageRepository: productImgRepo,
	}

}
