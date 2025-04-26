package api

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/models"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
	queries, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	users, err := queries.GetUsers(*ctx)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	sanitizedUsers, ok := models.SanitizeUsers(c, &users)
	if !ok {
		return nil
	}

	return status.Ok(c, &sanitizedUsers)
}

func GetUser(c *fiber.Ctx) error {
	queries, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	userId, err := c.ParamsInt("user_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid user_id"))
	}

	user, err := queries.GetUser(*ctx, int64(userId))
	if err != nil {
		return status.NotFound(c, nil)
	}

	sanitizedUser, ok := models.SanitizeUser(c, &user)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedUser)
}

func GetMe(c *fiber.Ctx) error {
	user, ok := getUser(c)
	if !ok {
		return status.Unauthorized(c, nil)
	}

	sanitizedUser, ok := models.SanitizeUser(c, user)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedUser)
}

func RemoveMe(c *fiber.Ctx) error {
	queries, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	_, err := queries.DeleteUser(ctx, user.ID)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, nil)
}

func UpdateMe(c *fiber.Ctx) error {
	queries, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	ok = schema.GetUpdateMeInput(c, &user)
	if !ok {
		return nil
	}

	newUser, err := queries.UpdateUser(ctx)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	sanitizedUser, ok := models.SanitizeUser(c, newUser)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedUser)
}
