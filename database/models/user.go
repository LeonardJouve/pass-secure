package models

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/fiber/v2"
)

type SanitizedUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func SanitizeUser(_ *fiber.Ctx, user *queries.User) SanitizedUser {
	return SanitizedUser{
		ID:    user.ID,
		Email: user.Email,
	}
}

func SanitizeUsers(c *fiber.Ctx, users *[]queries.User) []SanitizedUser {
	sanitizedUsers := make([]SanitizedUser, len(*users))
	for i, user := range *users {
		sanitizedUsers[i] = SanitizeUser(c, &user)
	}

	return sanitizedUsers
}
