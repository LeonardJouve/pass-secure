package schemas

import (
	"fmt"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/models"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type CreateFolderInput struct {
	Name     string `json:"name" validate:"required"`
	ParentID uint   `json:"parentId" validate:"required"`
}

func GetCreateFolderInput(c *fiber.Ctx) (models.Folder, bool) {
	var input CreateFolderInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return models.Folder{}, false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return models.Folder{}, false
	}

	return models.Folder{
		Name:     input.Name,
		ParentID: &input.ParentID,
	}, true
}

type UpdateFolderInput struct {
	Name     string `json:"name"`
	ParentID uint   `json:"parentId"`
}

func GetUpdateFolderInput(c *fiber.Ctx, folder *models.Folder) bool {
	var input UpdateFolderInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return false
	}

	if len(input.Name) > 0 {
		folder.Name = input.Name
	}

	if input.ParentID != 0 {
		folder.ParentID = &input.ParentID
	}

	return true
}

type InviteToFolderInput struct {
	Email  string   `json:"email"`
	Emails []string `json:"emails"`
}

func GetInviteToFolderInput(c *fiber.Ctx, folder *models.Folder) bool {
	var input InviteToFolderInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return false
	}

	if len(input.Email) > 0 {
		input.Emails = append(input.Emails, input.Email)
	}

	var user models.User
	for _, email := range input.Emails {
		if err := database.Database.Where("email = ?", email).First(&user).Error; err != nil {
			status.NotFound(c, fmt.Errorf("user \"%s\" could not be found", input.Email))
			return false
		}

		isAlreadyIn := false
		for _, u := range folder.Users {
			if u.ID == user.ID {
				isAlreadyIn = true
				break
			}
		}

		if isAlreadyIn {
			continue
		}

		folder.Users = append(folder.Users, user)
	}

	return true
}
