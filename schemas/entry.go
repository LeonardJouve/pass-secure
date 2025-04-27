package schemas

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type CreateEntryInput struct {
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Url      string `json:"url" validate:"omitempty"`
	FolderID int64  `json:"folderId" validate:"required"`
}

func GetCreateEntryInput(c *fiber.Ctx) (queries.CreateEntryParams, bool) {
	var input CreateEntryInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.CreateEntryParams{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.CreateEntryParams{}, false
	}

	result := queries.CreateEntryParams{
		Name:     input.Name,
		Username: input.Username,
		Password: input.Password,
		Url:      &input.Url,
		FolderID: input.FolderID,
	}

	if len(input.Url) == 0 {
		result.Url = nil
	}

	return result, true
}

type UpdateEntryInput struct {
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Url      string `json:"url" validate:"omitempty"`
	FolderID int64  `json:"folderId" validate:"required"`
}

func GetUpdateEntryInput(c *fiber.Ctx) (queries.UpdateEntryParams, bool) {
	var input UpdateEntryInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.UpdateEntryParams{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.UpdateEntryParams{}, false
	}

	result := queries.UpdateEntryParams{
		Name:     input.Name,
		Username: input.Username,
		Password: input.Password,
		Url:      &input.Url,
		FolderID: input.FolderID,
	}

	if len(input.Url) == 0 {
		result.Url = nil
	}

	return result, true
}
