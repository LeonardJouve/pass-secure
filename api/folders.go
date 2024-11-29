package api

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func CreateFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folder, ok := schema.GetCreateFolderInput(c)
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}
	folder.OwnerID = user.ID

	parentFolder, ok := getUserFolder(c, *folder.ParentID)
	if !ok {
		return nil
	}

	if parentFolder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	if ok := database.Execute(c, tx.Create(&folder).Error); !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Model(&user).Association("Folders").Append(&folder)); !ok {
		return nil
	}

	return c.Status(fiber.StatusCreated).JSON(folder.Sanitize())
}

func UpdateFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid folder_id"))
	}

	folder, ok := getUserFolder(c, uint(folderId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	if folder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	ok = schema.GetUpdateFolderInput(c, &folder)
	if !ok {
		return nil
	}

	if database.Database.Updates(&folder).Error != nil {
		return nil
	}

	return c.Status(fiber.StatusOK).JSON(folder.Sanitize())
}

func GetFolders(c *fiber.Ctx) error {
	folders, ok := getUserFolders(c)
	if !ok {
		return nil
	}

	var sanitizedFolders []model.SanitizedFolder
	for _, folder := range folders {
		sanitizedFolders = append(sanitizedFolders, *folder.Sanitize())
	}

	return status.Ok(c, &sanitizedFolders)
}

func GetFolder(c *fiber.Ctx) error {
	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid folder_id"))
	}

	folder, ok := getUserFolder(c, uint(folderId))
	if !ok {
		return nil
	}

	return status.Ok(c, folder.Sanitize())
}

func RemoveFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid folder_id"))
	}

	folder, ok := getUserFolder(c, uint(folderId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	if folder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	if ok := database.Execute(c, tx.Model(&user).Association("Folders").Delete(&folder)); !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Unscoped().Delete(&folder).Error); !ok {
		return nil
	}

	return status.Ok(c, nil)
}

func getUserFolders(c *fiber.Ctx) ([]model.Folder, bool) {
	user, ok := getUser(c)
	if !ok {
		return []model.Folder{}, false
	}

	if database.Database.Preload("Folders").First(&user).Error != nil {
		status.InternalServerError(c, nil)
		return []model.Folder{}, false
	}

	return user.Folders, true
}

func getUserFolder(c *fiber.Ctx, folderId uint) (model.Folder, bool) {
	folders, ok := getUserFolders(c)
	if !ok {
		return model.Folder{}, false
	}

	var folder model.Folder
	for _, f := range folders {
		if f.ID == folderId {
			folder = f
		}
	}

	if folder.ID == 0 {
		status.NotFound(c, nil)
		return model.Folder{}, false
	}

	return folder, true
}
