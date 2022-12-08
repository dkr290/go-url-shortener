package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dkr290/go-url-shortener/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}
	app := fiber.New()

	app.Use(logger.New())

	setupRoutes(app)

	err = app.Listen(os.Getenv("APP_PORT"))
	if err != nil {
		log.Fatal(err)
	}
}

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}
