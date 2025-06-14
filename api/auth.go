package api

import (
	"errors"
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
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

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

	userId, err := strconv.ParseInt(accessTokenClaims.Subject, 10, 64)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	user, err := qtx.GetUser(ctx, userId)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	c.Locals("user", user)

	return c.Next()
}

func Register(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	input, ok := schemas.GetRegisterUserInput(c)
	if !ok {
		return nil
	}

	exists, err := qtx.HasUserWithEmailOrUsername(ctx, queries.HasUserWithEmailOrUsernameParams{
		Email:    input.Email,
		Username: input.Username,
	})
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	if exists {
		return status.BadRequest(c, errors.New("user with same identifiers already exists"))
	}

	user, err := qtx.CreateUser(ctx, input)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Created(c, models.SanitizeUser(c, &user))
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
		"access_token": accessToken,
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

func GetCSRF(c *fiber.Ctx) error {
	csrfToken, ok := c.Locals(auth.CSRF_TOKEN).(string)
	if !ok || len(csrfToken) == 0 {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, fiber.Map{
		"csrf_token": csrfToken,
	})
}
