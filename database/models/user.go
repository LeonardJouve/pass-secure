package models

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type SanitizedUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func SanitizeUser(_ *fiber.Ctx, user *queries.User) (SanitizedUser, bool) {
	return SanitizedUser{
		ID:    user.ID,
		Email: user.Email,
	}, true
}

func SanitizeUsers(c *fiber.Ctx, users *[]queries.User) ([]SanitizedUser, bool) {
	sanitizedUsers := []SanitizedUser{}
	for _, user := range *users {
		sanitizedUser, ok := SanitizeUser(c, &user)
		if !ok {
			status.InternalServerError(c, nil)
			return []SanitizedUser{}, false
		}

		sanitizedUsers = append(sanitizedUsers, sanitizedUser)
	}

	return sanitizedUsers, true
}
