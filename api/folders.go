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
	"gorm.io/gorm"
)

func CreateFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folder, ok := schemas.GetCreateFolderInput(c)
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	parentFolder, ok := getUserFolder(c, *folder.ParentID)
	if !ok {
		return nil
	}

	if err := createFolder(c, tx, &folder, &user, &parentFolder); err != nil {
		return nil
	}

	return status.Created(c, folder.Sanitize())
}

func UpdateFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
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

	ok = schemas.GetUpdateFolderInput(c, &folder)
	if !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Updates(&folder).Error); !ok {
		return nil
	}

	return status.Ok(c, folder.Sanitize())
}

func GetFolders(c *fiber.Ctx) error {
	folders, ok := getUserFolders(c)
	if !ok {
		return nil
	}

	sanitizedFolders := []models.SanitizedFolder{}
	for _, folder := range folders {
		sanitizedFolders = append(sanitizedFolders, *folder.Sanitize())
	}

	return status.Ok(c, &sanitizedFolders)
}

func GetFolder(c *fiber.Ctx) error {
	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
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
		return status.BadRequest(c, errors.New("invalid folder_id"))
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

	if folder.ParentID == nil {
		return status.Unauthorized(c, nil)
	}

	if ok := database.Execute(c, tx.Delete(&folder).Error); !ok {
		return nil
	}

	return status.Ok(c, nil)
}

func InviteToFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
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

	if folder.ParentID == nil {
		return status.Unauthorized(c, nil)
	}

	if ok := schemas.GetInviteToFolderInput(c, &folder); !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Model(&folder).Association("Users").Replace(folder.Users)); !ok {
		return nil
	}

	return status.Ok(c, folder.Sanitize())
}

func RemoveInviteToFolder(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	folderId, err := c.ParamsInt("folder_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid folder_id"))
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

	inviteUserId, err := c.ParamsInt("user_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid user_id"))
	}

	found := false
	for _, user := range folder.Users {
		if user.ID == uint(inviteUserId) {
			found = true
		}
	}
	if !found {
		status.BadRequest(c, errors.New("invalid user_id"))
	}

	var inviteUser models.User
	if ok := database.Execute(c, tx.First(&inviteUser, inviteUserId).Error); !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Model(&folder).Association("Users").Delete(&inviteUser)); !ok {
		return nil
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

func createFolder(c *fiber.Ctx, qtx *queries.Queries, folder *queries.Folder, user *queries.User, parent *queries.Folder) error {
	if parent == nil {
		err := tx.Where("parent_id IS NULL AND owner_id = ?", user.ID).First(&models.Folder{}).Error
		if err == nil {
			return status.Unauthorized(c, nil)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return status.InternalServerError(c, nil)
		}
	} else if parent.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	if ok := database.Execute(c, tx.Create(folder).Error); !ok {
		return status.InternalServerError(c, nil)
	}

	if ok := database.Execute(c, tx.Model(folder).Association("Users").Append(user)); !ok {
		return status.InternalServerError(c, nil)
	}

	if ok := database.Execute(c, tx.Model(folder).Association("Owner").Append(user)); !ok {
		return status.InternalServerError(c, nil)
	}

	if parent != nil {
		if ok := database.Execute(c, tx.Model(folder).Association("Parent").Append(parent)); !ok {
			return status.InternalServerError(c, nil)
		}
	}

	return nil
}
