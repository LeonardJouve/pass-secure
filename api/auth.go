package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/pass-secure/auth"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func Protect(c *fiber.Ctx) error {
	var accessToken string
	authorization := c.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken = strings.TrimPrefix(authorization, "Bearer ")
	} else if accessTokenCookie := c.Cookies(auth.ACCESS_TOKEN); len(accessTokenCookie) != 0 {
		accessToken = accessTokenCookie
	}

	accessTokenClaims, ok := auth.ValidateToken(c, accessToken)
	if !ok {
		return nil
	}

	expired, ok := auth.IsExpired(c, accessTokenClaims)
	if !ok || expired {
		return nil
	}

	userId, err := strconv.ParseUint(accessTokenClaims.Subject, 10, 64)
	if err != nil {
		status.InternalServerError(c, nil)
		return nil
	}

	var user model.User
	if err := database.Database.First(&user, userId).Error; err != nil {
		return status.InternalServerError(c, nil)
	}
	if user.ID == 0 {
		return status.Unauthorized(c, nil)
	}

	c.Locals("user", user)

	return c.Next()
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

func getUser(c *fiber.Ctx) (model.User, bool) {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		status.InternalServerError(c, nil)
		return model.User{}, false
	}

	return user, true
}
