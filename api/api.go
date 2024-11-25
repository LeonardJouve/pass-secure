package api

import (
	"github.com/LeonardJouve/pass-secure/auth"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return status.Ok(c, nil)
}

func Register(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	user, ok := schema.GetRegisterUserInput(c)
	if !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Create(&user).Error); !ok {
		return nil
	}

	return status.Created(c, user.Sanitize())
}

func Login(c *fiber.Ctx) error {
	user, ok := schema.GetLoginUserInput(c)
	if !ok {
		return nil
	}

	accessToken, ok := auth.CreateToken(c, user.ID)
	if !ok {
		return nil
	}

	return status.Ok(c, fiber.Map{
		"accessToken": accessToken,
	})
}
