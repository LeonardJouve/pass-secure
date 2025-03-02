package schema

import (
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type UpdateMeInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GetUpdateMeInput(c *fiber.Ctx, user *model.User) bool {
	var input UpdateMeInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return false
	}

	if len(input.Email) > 0 {
		user.Email = input.Email
	}

	if len(input.Password) > 0 {
		user.Password = input.Password
	}

	return true
}
