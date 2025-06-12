package schemas

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type CreateFolderInput struct {
	Name     string `json:"name" validate:"required"`
	ParentID int64  `json:"parentId" validate:"required"`
}

func GetCreateFolderInput(c *fiber.Ctx, userId int64) (queries.CreateFolderParams, bool) {
	var input CreateFolderInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.CreateFolderParams{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.CreateFolderParams{}, false
	}

	return queries.CreateFolderParams{
		Name:     input.Name,
		OwnerID:  userId,
		ParentID: &input.ParentID,
	}, true
}

type UpdateFolderInput struct {
	Name     string `json:"name" validate:"required"`
	OwnerID  int64  `json:"ownerId" validate:"required"`
	ParentID int64  `json:"parentId" validate:"required"`
}

func GetUpdateFolderInput(c *fiber.Ctx) (queries.UpdateFolderParams, bool) {
	var input UpdateFolderInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return queries.UpdateFolderParams{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return queries.UpdateFolderParams{}, false
	}

	return queries.UpdateFolderParams{
		Name:     input.Name,
		OwnerID:  input.OwnerID,
		ParentID: &input.ParentID,
	}, true
}

type AddFolderUserInput struct {
	Email string `json:"email" validate:"required"`
}

func GetAddFolderUserInput(c *fiber.Ctx) (AddFolderUserInput, bool) {
	var input AddFolderUserInput
	if err := c.BodyParser(&input); err != nil {
		status.BadRequest(c, err)
		return AddFolderUserInput{}, false
	}

	if err := validate.Struct(input); err != nil {
		status.BadRequest(c, err)
		return AddFolderUserInput{}, false
	}

	return input, true
}
