package configs

import (
	"backEnd-RingoTechLife/internal/storage"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Router  chi.Router
	Repo    *RepositoryConfigs
	Service *ServiceConfigs
}

func NewApp(ctx context.Context, dbString string, r chi.Router) *App {

	pool, err := setUpDatabase(ctx, dbString)
	if err != nil {
		panic(err)
	}

	serverStorage := storage.NewFileServerStorage()
	repoCfg := NewRepositoryConfigs(pool)
	serviceCfg := NewServiceConfigs(repoCfg, serverStorage)

	SetupRouter(r, serviceCfg)

	return &App{
		Router:  r,
		Repo:    repoCfg,
		Service: serviceCfg,
	}
}

func (a *App) Run() {
	fmt.Println("Server berjalan di port 80")
	http.ListenAndServe(":80", a.Router)

}
