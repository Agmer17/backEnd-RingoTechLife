package main

import (
	"backEnd-RingoTechLife/configs"
	"backEnd-RingoTechLife/pkg"
	"context"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	app := configs.NewApp(mainContext, dbUrl, router)

	app.Run()

}
