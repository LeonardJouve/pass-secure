package schemas

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type UpdateMeInput struct {
	Email    string `json:"email" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func GetUpdateMeInput(c *fiber.Ctx, id int64) (queries.UpdateUserParams, bool) {
	var input UpdateMeInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.UpdateUserParams{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.UpdateUserParams{}, false
	}

	return queries.UpdateUserParams{
		ID:       id,
		Email:    input.Email,
		Username: input.Username,
		Password: input.Password,
	}, true
}
