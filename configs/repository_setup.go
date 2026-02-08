package configs

import (
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/products"
	"backEnd-RingoTechLife/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryConfigs struct {
	UserRepository     *user.UserRepositoryImpl
	CategoryRepository *category.CategoryRepositoryImpl
	ProductsRepository *products.ProductRepositoryImpl
}

func NewRepositoryConfigs(pool *pgxpool.Pool) *RepositoryConfigs {

	userRepo := user.NewUserRepository(pool)
	categoryRepo := category.NewCategoryRepository(pool)
	productRepo := products.NewProductsRepository(pool)

	return &RepositoryConfigs{
		UserRepository:     userRepo,
		CategoryRepository: categoryRepo,
		ProductsRepository: productRepo,
	}

}
