package schema

import (
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type CreateEntryInput struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	FolderID uint   `json:"folderId" validate:"required"`
}

func GetCreateEntryInput(c *fiber.Ctx) (model.Entry, bool) {
	var input CreateEntryInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return model.Entry{}, false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return model.Entry{}, false
	}

	return model.Entry{
		Name:     input.Name,
		Password: input.Password,
		FolderID: input.FolderID,
	}, true
}

type UpdateEntryInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	FolderID uint   `json:"folderId"`
}

func GetUpdateEntryInput(c *fiber.Ctx, entry *model.Entry) bool {
	var input UpdateEntryInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return false
	}

	if len(input.Name) > 0 {
		entry.Name = input.Name
	}

	if len(input.Password) > 0 {
		entry.Password = input.Password
	}

	if input.FolderID != 0 {
		entry.FolderID = input.FolderID
	}

	return true
}
