package main

import (
	"backEnd-RingoTechLife/configs"
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
	// jwtSecret := os.Getenv("JWT_SECRET")

	router := chi.NewRouter()

	app := configs.NewApp(dbUrl, router)

	app.Run()

}
