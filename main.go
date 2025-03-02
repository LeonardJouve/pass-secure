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

	apiGroup := app.Group("", api.Protect)

	folderGroup := apiGroup.Group("/folders")
	folderGroup.Get("/", api.GetFolders)
	folderGroup.Get("/:folder_id", api.GetFolder)
	folderGroup.Post("/", api.CreateFolder)
	folderGroup.Put("/:folder_id", api.UpdateFolder)
	folderGroup.Delete("/:folder_id", api.RemoveFolder)
	// TODO: add user

	entriesGroup := apiGroup.Group("/entries")
	entriesGroup.Get("/", api.GetEntries)
	entriesGroup.Get("/:entry_id", api.GetEntry)
	entriesGroup.Post("/", api.CreateEntry)
	entriesGroup.Delete("/:entry_id", api.RemoveEntry)
	// TODO: update entry

	usersGroup := apiGroup.Group("/users")
	usersGroup.Get("/", api.GetUsers)
	usersGroup.Get("/:user_id", api.GetUser)
	// TODO: remove user
	// TODO: update user
	// TODO: get me

	app.Listen(":3000")
}
