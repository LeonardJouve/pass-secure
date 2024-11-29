package main

import (
	"os"
	"path/filepath"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/gofiber/fiber/v2"
)

func main() {
	schema.Init()

	path, err := os.Executable()
	if err != nil {
		panic("Could not retrieve executable path")
	}
	databasePath := filepath.Join(filepath.Dir(path), "..", "database.db")
	err = database.Init(databasePath)
	if err != nil {
		panic("Could not initialize database")
	}

	model.Migrate()

	app := fiber.New()

	app.Get("/healthcheck", api.HealthCheck)

	app.Post("/login", api.Login)
	app.Post("/register", api.Register)

	apiGroup := app.Group("/api", api.Protect)

	entriesGroup := apiGroup.Group("/entries")
	entriesGroup.Get("/", api.GetEntries)

	app.Listen(":3000")
}
