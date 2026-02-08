package configs

import (
	"backEnd-RingoTechLife/internal/auth"
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/storage"
	"backEnd-RingoTechLife/internal/user"
)

type ServiceConfigs struct {
	AuthService     *auth.AuthService
	UserService     *user.UserService
	ServerStorage   *storage.FileStorage
	CategoryService *category.CategoryService
}

func NewServiceConfigs(rcf *RepositoryConfigs, serverStorage *storage.FileStorage) *ServiceConfigs {

	userSvc := user.NewUserService(rcf.UserRepository, serverStorage)
	authSvc := auth.NewAuthService(userSvc)
	categorySvc := category.NewCategoryService(rcf.CategoryRepository)

	return &ServiceConfigs{
		AuthService:     authSvc,
		UserService:     userSvc,
		ServerStorage:   serverStorage,
		CategoryService: categorySvc,
	}

}
