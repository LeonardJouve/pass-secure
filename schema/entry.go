package schema

import (
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type CreateEntryInput struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	ParentID uint   `json:"parentId" validate:"required"`
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
		ParentID: input.ParentID,
	}, true
}

type UpdateEntryInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	ParentID uint   `json:"parentId"`
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

	entry.Name = input.Name
	entry.Password = input.Password
	entry.ParentID = input.ParentID

	return true
}
