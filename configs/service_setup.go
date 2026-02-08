package configs

import (
	"backEnd-RingoTechLife/internal/auth"
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/products"
	"backEnd-RingoTechLife/internal/storage"
	"backEnd-RingoTechLife/internal/user"
)

type ServiceConfigs struct {
	AuthService     *auth.AuthService
	UserService     *user.UserService
	ServerStorage   *storage.FileStorage
	CategoryService *category.CategoryService
	ProductService  *products.ProductsService
}

func NewServiceConfigs(rcf *RepositoryConfigs, serverStorage *storage.FileStorage) *ServiceConfigs {

	userSvc := user.NewUserService(rcf.UserRepository, serverStorage)
	authSvc := auth.NewAuthService(userSvc)
	categorySvc := category.NewCategoryService(rcf.CategoryRepository)
	productSvc := products.NewProductsService(rcf.ProductsRepository)

	return &ServiceConfigs{
		AuthService:     authSvc,
		UserService:     userSvc,
		ServerStorage:   serverStorage,
		CategoryService: categorySvc,
		ProductService:  productSvc,
	}

}
