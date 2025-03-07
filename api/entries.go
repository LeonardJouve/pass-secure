package api

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func CreateEntry(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	entry, ok := schema.GetCreateEntryInput(c)
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	parentFolder, ok := getUserFolder(c, entry.FolderID)
	if !ok {
		return nil
	}

	if parentFolder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	if ok := database.Execute(c, tx.Create(&entry).Error); !ok {
		return nil
	}

	if ok := database.Execute(c, tx.Model(&entry).Association("Folder").Replace(&parentFolder)); !ok {
		return nil
	}

	return status.Created(c, entry.Sanitize())
}

func GetEntries(c *fiber.Ctx) error {
	entries, ok := getUserEntries(c)
	if !ok {
		return nil
	}

	sanitizedEntries := []model.SanitizedEntry{}
	for _, entry := range entries {
		sanitizedEntries = append(sanitizedEntries, *entry.Sanitize())
	}

	return status.Ok(c, &sanitizedEntries)
}

func GetEntry(c *fiber.Ctx) error {
	entryId, err := c.ParamsInt("entry_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid entry_id"))
	}

	entry, ok := getUserEntry(c, uint(entryId))
	if !ok {
		return nil
	}

	return status.Ok(c, entry.Sanitize())
}

func UpdateEntry(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	entryId, err := c.ParamsInt("entry_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid entry_id"))
	}

	entry, ok := getUserEntry(c, uint(entryId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	if entry.Folder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	ok = schema.GetUpdateEntryInput(c, &entry)
	if !ok {
		return nil
	}

	if database.Database.Updates(&entry).Error != nil {
		return nil
	}

	return status.Ok(c, entry.Sanitize())
}

func RemoveEntry(c *fiber.Ctx) error {
	entryId, err := c.ParamsInt("entry_id")
	if err != nil {
		status.BadRequest(c, errors.New("invalid entry_id"))
	}

	entry, ok := getUserEntry(c, uint(entryId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	if entry.Folder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	if database.Database.Delete(&entry).Error != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, nil)
}

func getUserEntries(c *fiber.Ctx) ([]model.Entry, bool) {
	folders, ok := getUserFolders(c)
	if !ok {
		return []model.Entry{}, false
	}

	entries := []model.Entry{}
	for _, folder := range folders {
		for _, entry := range folder.Entries {
			if err := database.Database.Model(&entry).Association("Folder").Find(&entry.Folder); err != nil {
				status.InternalServerError(c, nil)
				return []model.Entry{}, false
			}
			entries = append(entries, entry)
		}
	}

	return entries, true
}

func getUserEntry(c *fiber.Ctx, entryId uint) (model.Entry, bool) {
	entries, ok := getUserEntries(c)
	if !ok {
		return model.Entry{}, false
	}

	var entry model.Entry
	for _, e := range entries {
		if e.ID == entryId {
			entry = e
			break
		}
	}

	if entry.ID == 0 {
		status.NotFound(c, nil)
		return model.Entry{}, false
	}

	return entry, true
}
