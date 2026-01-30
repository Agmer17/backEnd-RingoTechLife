package configs

import (
	"context"
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

	repoCfg := NewRepositoryConfigs(pool)
	serviceCfg := NewServiceConfigs(repoCfg)

	SetupRouter(r, serviceCfg)

	return &App{
		Router:  r,
		Repo:    repoCfg,
		Service: serviceCfg,
	}
}

func (a *App) Run() {
	http.ListenAndServe(":80", a.Router)
}
