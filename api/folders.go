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

	folder, ok := createFolder(c, &input, &user, &parentFolder)
	if !ok {
		return nil
	}

	return status.Created(c, models.SanitizeFolder(c, &folder))
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

	return status.Ok(c, models.SanitizeFolder(c, &newFolder))
}

func GetFolders(c *fiber.Ctx) error {
	folders, ok := getUserFolders(c)
	if !ok {
		return nil
	}

	return status.Ok(c, models.SanitizeFolders(c, &folders))
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

	return status.Ok(c, models.SanitizeFolder(c, &folder))
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

func createFolder(c *fiber.Ctx, input *queries.CreateFolderParams, user *queries.User, parent *queries.Folder) (queries.Folder, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return queries.Folder{}, false
	}
	defer commit()

	if parent == nil {
		// TODO: do not allow nil parent as it is created with the user
		_, err := qtx.GetUserRootFolder(*ctx, user.ID)
		if err == nil {
			status.BadRequest(c, errors.New("invalid parent_id"))
			return queries.Folder{}, false
		} else if !errors.Is(err, sql.ErrNoRows) {
			status.InternalServerError(c, nil)
			return queries.Folder{}, false
		}
	} else if parent.OwnerID != user.ID {
		status.Unauthorized(c, nil)
		return queries.Folder{}, false
	}

	folder, err := qtx.CreateFolder(*ctx, *input)
	if err != nil {
		status.InternalServerError(c, nil)
		return queries.Folder{}, false
	}

	return folder, true
}
