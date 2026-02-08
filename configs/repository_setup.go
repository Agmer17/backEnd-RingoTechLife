package configs

import (
	"backEnd-RingoTechLife/internal/category"
	"backEnd-RingoTechLife/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryConfigs struct {
	UserRepository     *user.UserRepositoryImpl
	CategoryRepository *category.CategoryRepositoryImpl
}

func NewRepositoryConfigs(pool *pgxpool.Pool) *RepositoryConfigs {

	userRepo := user.NewUserRepository(pool)
	categoryRepo := category.NewCategoryRepository(pool)

	return &RepositoryConfigs{
		UserRepository:     userRepo,
		CategoryRepository: categoryRepo,
	}

}
