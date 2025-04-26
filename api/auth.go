package api

import (
	"strconv"
	"strings"

	"github.com/LeonardJouve/pass-secure/auth"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/models"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/schemas"
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

	expired := auth.IsExpired(c, accessTokenClaims)
	if expired {
		return nil
	}

	userId, err := strconv.ParseUint(accessTokenClaims.Subject, 10, 64)
	if err != nil {
		status.InternalServerError(c, nil)
		return nil
	}

	var user models.User
	if err := database.Database.First(&user, userId).Error; err != nil {
		return status.InternalServerError(c, nil)
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

	user, ok := schemas.GetRegisterUserInput(c)
	if !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Create(&user).Error); !ok {
		return nil
	}

	folder := models.Folder{
		Name: "",
	}
	if err := createFolder(c, tx, &folder, &user, nil); err != nil {
		return nil
	}

	return status.Created(c, user.Sanitize())
}

func Login(c *fiber.Ctx) error {
	user, ok := schemas.GetLoginUserInput(c)
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

func getUser(c *fiber.Ctx) (queries.User, bool) {
	user, ok := c.Locals("user").(queries.User)
	if !ok {
		status.InternalServerError(c, nil)
		return queries.User{}, false
	}

	return user, true
}
