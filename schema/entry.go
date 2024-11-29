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
