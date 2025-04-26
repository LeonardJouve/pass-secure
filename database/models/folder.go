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
	ParentID int64   `json:"parentId"`
}

func SanitizeFolder(c *fiber.Ctx, folder *queries.Folder) (SanitizedFolder, bool) {
	db, err := database.GetInstance()
	if err != nil {
		status.InternalServerError(c, nil)
		return SanitizedFolder{}, false
	}

	qtx, ctx, commit, ok := db.BeginTransaction(c)
	if !ok {
		return SanitizedFolder{}, false
	}
	defer commit()

	userIds, err := qtx.GetFolderUserIds(*ctx, folder.ID)
	if err != nil {
		return SanitizedFolder{}, false
	}

	return SanitizedFolder{
		ID:       folder.ID,
		UserIds:  userIds,
		OwnerID:  folder.OwnerID,
		Name:     folder.Name,
		ParentID: folder.ParentID.Int64,
	}, true
}
