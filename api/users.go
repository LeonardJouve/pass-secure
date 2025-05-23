package api

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/models"
	"github.com/LeonardJouve/pass-secure/schemas"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	users, err := qtx.GetUsers(ctx)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, models.SanitizeUsers(c, &users))
}

func GetUser(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	userId, err := c.ParamsInt("user_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid user_id"))
	}

	user, err := qtx.GetUser(ctx, int64(userId))
	if err != nil {
		return status.NotFound(c, nil)
	}

	return status.Ok(c, models.SanitizeUser(c, &user))
}

func GetMe(c *fiber.Ctx) error {
	user, ok := getUser(c)
	if !ok {
		return status.Unauthorized(c, nil)
	}

	return status.Ok(c, models.SanitizeUser(c, &user))
}

func RemoveMe(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	err := qtx.DeleteUser(ctx, user.ID)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, nil)
}

func UpdateMe(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	input, ok := schemas.GetUpdateMeInput(c, user.ID)
	if !ok {
		return nil
	}

	newUser, err := qtx.UpdateUser(ctx, input)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, models.SanitizeUser(c, &newUser))
}
