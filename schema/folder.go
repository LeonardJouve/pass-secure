package schema

import (
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type CreateFolderInput struct {
	Name     string `json:"name" validate:"required"`
	ParentID uint   `json:"parentId" validate:"required"`
}

func GetCreateFolderInput(c *fiber.Ctx) (model.Folder, bool) {
	var input CreateFolderInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return model.Folder{}, false
	}
	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return model.Folder{}, false
	}

	return model.Folder{
		Name:     input.Name,
		ParentID: &input.ParentID,
	}, true
}

type UpdateFolderInput struct {
	Name     string `json:"name"`
	ParentID uint   `json:"parentId"`
}

func GetUpdateFolderInput(c *fiber.Ctx, folder *model.Folder) bool {
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
