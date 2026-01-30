package main

import (
	"backEnd-RingoTechLife/configs"
	"backEnd-RingoTechLife/pkg"
	"context"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	dbUrl := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	mainContext := context.Background()

	pkg.JwtInit(jwtSecret)
	router := chi.NewRouter()

	app := configs.NewApp(mainContext, dbUrl, router)

	app.Run()

}
