package main

import (
	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/gofiber/fiber/v2"
)

func main() {
	err := database.Init("./db.db")
	if err != nil {
		panic("Could not initialize database")
	}

	app := fiber.New()

	app.Get("/healthcheck", api.HealthCheck)

	app.Post("/login", api.Login)
	app.Post("/register", api.Register)

	app.Listen(":3000")
}
