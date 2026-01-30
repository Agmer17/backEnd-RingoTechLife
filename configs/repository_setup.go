package configs

import (
	"backEnd-RingoTechLife/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryConfigs struct {
	UserRepository *user.UserRepositoryImpl
}

func NewRepositoryConfigs(pool *pgxpool.Pool) *RepositoryConfigs {

	userRepo := user.NewUserRepository(pool)

	return &RepositoryConfigs{
		UserRepository: userRepo,
	}

}
