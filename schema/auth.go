package schema

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterInput struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,min=8"`
}

func GetRegisterUserInput(c *fiber.Ctx) (model.User, bool) {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return model.User{}, false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return model.User{}, false
	}

	if input.Password != input.PasswordConfirm {
		status.BadRequest(c, errors.New("invalid password confirmation"))
		return model.User{}, false
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		status.InternalServerError(c, nil)
		return model.User{}, false
	}

	return model.User{
		Email:    input.Email,
		Password: string(hashedPassword),
	}, true
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func GetLoginUserInput(c *fiber.Ctx) (model.User, bool) {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return model.User{}, false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return model.User{}, false
	}

	var user model.User
	if err := database.Database.Where(&model.User{Email: input.Email}).First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		status.InternalServerError(c, nil)
		return model.User{}, false
	}

	if user.ID == 0 {
		status.Unauthorized(c, errors.New("invalid credentials"))
		return model.User{}, false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		status.Unauthorized(c, errors.New("invalid credentials"))
		return model.User{}, false
	}

	return user, true
}
