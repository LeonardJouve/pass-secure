package schemas

import (
	"database/sql"
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

func GetRegisterUserInput(c *fiber.Ctx) (queries.CreateUserParams, bool) {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.CreateUserParams{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.CreateUserParams{}, false
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		status.InternalServerError(c, nil)
		return queries.CreateUserParams{}, false
	}

	return queries.CreateUserParams{
		Email:    input.Email,
		Username: input.Username,
		Password: string(hashedPassword),
	}, true
}

type LoginInput struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func GetLoginUserInput(c *fiber.Ctx) (queries.User, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return queries.User{}, false
	}
	defer commit()

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.User{}, false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.User{}, false
	}

	invalidCredentialsErr := errors.New("invalid credentials")

	user, err := qtx.GetUserByEmail(*ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status.Unauthorized(c, invalidCredentialsErr)
			return queries.User{}, false
		} else {
			status.InternalServerError(c, nil)
			return queries.User{}, false
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		status.Unauthorized(c, invalidCredentialsErr)
		return queries.User{}, false
	}

	return user, true
}
