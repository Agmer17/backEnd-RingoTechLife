package configs

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DbPool  *pgxpool.Pool
	Router  *chi.Mux
	Repo    *RepositoryConfigs
	Service *ServiceConfigs
}

func NewApp(dbString string, r *chi.Mux) *App {
	return &App{}
}

func (a *App) Run() {
	http.ListenAndServe(":80", a.Router)
}
