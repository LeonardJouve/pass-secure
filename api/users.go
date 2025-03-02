package api

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
	users := []model.User{}
	if database.Database.Find(&users).Error != nil {
		status.InternalServerError(c, nil)
		return status.InternalServerError(c, nil)
	}

	sanitizedUsers := []model.SanitizedUser{}
	for _, user := range users {
		sanitizedUsers = append(sanitizedUsers, *user.Sanitize())
	}

	return status.Ok(c, &sanitizedUsers)
}

func GetUser(c *fiber.Ctx) error {
	userId, err := c.ParamsInt("user_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid user_id"))
	}

	var user model.User
	if err := database.Database.First(&user, userId).Error; err != nil {
		return status.NotFound(c, nil)
	}

	return status.Ok(c, user.Sanitize())
}

func GetMe(c *fiber.Ctx) error {
	user, ok := getUser(c)
	if !ok {
		return status.Unauthorized(c, nil)
	}

	return status.Ok(c, user.Sanitize())
}
