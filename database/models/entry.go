package models

import (
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/gofiber/fiber/v2"
)

type SanitizedEntry struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	FolderID int64  `json:"folderId"`
}

func SanitizeEntry(_ *fiber.Ctx, entry *queries.Entry) (SanitizedEntry, bool) {
	return SanitizedEntry{
		ID:       entry.ID,
		Name:     entry.Name,
		Password: entry.Password,
		FolderID: entry.FolderID,
	}, true
}
