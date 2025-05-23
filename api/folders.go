package api

import (
	"database/sql"
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/models"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/schemas"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func CreateFolder(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return status.InternalServerError(c, nil)
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	input, ok := schemas.GetCreateFolderInput(c, user.ID)
	if !ok {
		return nil
	}

	parentFolder, ok := getUserFolder(c, *input.ParentID)
	if !ok {
		return nil
	}

	if parentFolder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	folder, err := qtx.CreateFolder(*ctx, input)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	sanitizedFolder, ok := models.SanitizeFolder(c, &folder)
	if !ok {
		return nil
	}

	return status.Created(c, sanitizedFolder)
}

func UpdateFolder(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
	}

	folder, ok := getUserFolder(c, int64(folderId))
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

	input, ok := schemas.GetUpdateFolderInput(c)
	if !ok {
		return nil
	}

	newFolder, err := qtx.UpdateFolder(*ctx, input)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	sanitizedFolder, ok := models.SanitizeFolder(c, &newFolder)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedFolder)
}

func GetFolders(c *fiber.Ctx) error {
	folders, ok := getUserFolders(c)
	if !ok {
		return nil
	}

	sanitizedFolders, ok := models.SanitizeFolders(c, &folders)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedFolders)
}

func GetFolder(c *fiber.Ctx) error {
	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
	}

	folder, ok := getUserFolder(c, int64(folderId))
	if !ok {
		return nil
	}

	sanitizedFolder, ok := models.SanitizeFolder(c, &folder)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedFolder)
}

func RemoveFolder(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
	}

	folder, ok := getUserFolder(c, int64(folderId))
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

	if folder.ParentID == nil {
		return status.Unauthorized(c, nil)
	}

	err = qtx.DeleteFolder(*ctx, folder.ID)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, nil)
}

func getUserFolders(c *fiber.Ctx) ([]queries.Folder, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return []queries.Folder{}, false
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return []queries.Folder{}, false
	}

	folders, err := qtx.GetUserFolders(*ctx, user.ID)
	if err != nil {
		status.InternalServerError(c, nil)
		return []queries.Folder{}, false
	}

	return folders, true
}

func getUserFolder(c *fiber.Ctx, folderId int64) (queries.Folder, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return queries.Folder{}, false
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return queries.Folder{}, false
	}

	folder, err := qtx.GetUserFolder(*ctx, queries.GetUserFolderParams{
		UserID:   user.ID,
		FolderID: folderId,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status.NotFound(c, nil)
		} else {
			status.InternalServerError(c, nil)
		}

		return queries.Folder{}, false
	}

	return folder, true
}
