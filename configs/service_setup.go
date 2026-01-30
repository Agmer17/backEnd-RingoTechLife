package configs

import (
	"backEnd-RingoTechLife/internal/auth"
	"backEnd-RingoTechLife/internal/user"
)

type ServiceConfigs struct {
	AuthService *auth.AuthService
	UserService *user.UserService
}

func NewServiceConfigs(rcf *RepositoryConfigs) *ServiceConfigs {

	userSvc := user.NewUserService(rcf.UserRepository)
	authSvc := auth.NewAuthService(userSvc)

	return &ServiceConfigs{
		AuthService: authSvc,
		UserService: userSvc,
	}

}
