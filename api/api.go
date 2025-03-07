package api

import (
	"fmt"

	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return status.Ok(c, nil)
}

func Start(port uint16) func() error {
	app := fiber.New()

	app.Get("/healthcheck", HealthCheck)

	app.Post("/login", Login)
	app.Post("/register", Register)

	apiGroup := app.Group("", Protect)

	folderGroup := apiGroup.Group("/folders")
	folderGroup.Get("/", GetFolders)
	folderGroup.Get("/:folder_id", GetFolder)
	folderGroup.Post("/", CreateFolder)
	folderGroup.Put("/:folder_id/invite", InviteToFolder)
	folderGroup.Delete("/:folder_id/invite/:user_id", RemoveInviteToFolder)
	folderGroup.Put("/:folder_id", UpdateFolder)
	folderGroup.Delete("/:folder_id", RemoveFolder)

	entriesGroup := apiGroup.Group("/entries")
	entriesGroup.Get("/", GetEntries)
	entriesGroup.Get("/:entry_id", GetEntry)
	entriesGroup.Post("/", CreateEntry)
	entriesGroup.Put("/:entry_id", UpdateEntry)
	entriesGroup.Delete("/:entry_id", RemoveEntry)

	usersGroup := apiGroup.Group("/users")
	usersGroup.Get("/", GetUsers)
	usersGroup.Get("/me", GetMe)
	usersGroup.Delete("/me", RemoveMe)
	usersGroup.Put("/me", UpdateMe)
	usersGroup.Get("/:user_id", GetUser)

	app.Listen(fmt.Sprintf(":%d", port))

	return app.Shutdown
}
