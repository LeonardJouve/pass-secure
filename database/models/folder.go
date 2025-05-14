package models

import (
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

type SanitizedFolder struct {
	ID       int64   `json:"id"`
	UserIds  []int64 `json:"userIds"`
	OwnerID  int64   `json:"ownerId"`
	Name     string  `json:"name"`
	ParentID *int64  `json:"parentId"`
}

func SanitizeFolder(c *fiber.Ctx, folder *queries.Folder) (SanitizedFolder, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		status.InternalServerError(c, nil)
		return SanitizedFolder{}, false
	}
	defer commit()

	userIds, err := qtx.GetFolderUsers(*ctx, folder.ID)
	if err != nil {
		status.InternalServerError(c, nil)
		return SanitizedFolder{}, false
	}

	return SanitizedFolder{
		ID:       folder.ID,
		OwnerID:  folder.OwnerID,
		Name:     folder.Name,
		ParentID: folder.ParentID,
		UserIds:  userIds,
	}, true
}

func SanitizeFolders(c *fiber.Ctx, folders *[]queries.Folder) ([]SanitizedFolder, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		status.InternalServerError(c, nil)
		return []SanitizedFolder{}, false
	}
	defer commit()

	folderIds := make([]int64, len(*folders))
	for i, folder := range *folders {
		folderIds[i] = folder.ID
	}

	foldersUsers, err := qtx.GetFoldersUsers(*ctx, folderIds)
	if err != nil {
		status.InternalServerError(c, nil)
		return []SanitizedFolder{}, false
	}

	usersByFolder := make(map[int64][]int64)
	for _, folderUser := range foldersUsers {
		usersByFolder[folderUser.FolderID] = append(usersByFolder[folderUser.FolderID], folderUser.UserID)
	}

	sanitizedFolders := make([]SanitizedFolder, len(*folders))
	for i, folder := range *folders {
		sanitizedFolders[i] = SanitizedFolder{
			ID:       folder.ID,
			OwnerID:  folder.OwnerID,
			Name:     folder.Name,
			ParentID: folder.ParentID,
		}

		if userIds, ok := usersByFolder[folder.ID]; ok {
			sanitizedFolders[i].UserIds = userIds
		}
	}

	return sanitizedFolders, true
}
