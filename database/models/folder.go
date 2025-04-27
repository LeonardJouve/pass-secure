package models

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/fiber/v2"
)

type SanitizedFolder struct {
	ID       int64   `json:"id"`
	UserIds  []int64 `json:"userIds"`
	OwnerID  int64   `json:"ownerId"`
	Name     string  `json:"name"`
	ParentID *int64  `json:"parentId"`
}

func SanitizeFolder(c *fiber.Ctx, folder *queries.Folder) SanitizedFolder {
	return SanitizedFolder{
		ID:       folder.ID,
		OwnerID:  folder.OwnerID,
		Name:     folder.Name,
		ParentID: folder.ParentID,
	}
}

func SanitizeFolders(c *fiber.Ctx, folders *[]queries.Folder) []SanitizedFolder {
	sanitizedFolders := make([]SanitizedFolder, len(*folders))
	for i, folder := range *folders {
		sanitizedFolders[i] = SanitizeFolder(c, &folder)
	}

	return sanitizedFolders
}
