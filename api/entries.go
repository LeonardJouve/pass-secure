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

func CreateEntry(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	input, ok := schemas.GetCreateEntryInput(c)
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	parentFolder, ok := getUserFolder(c, input.FolderID)
	if !ok {
		return nil
	}

	if parentFolder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	entry, err := qtx.CreateEntry(*ctx, input)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	sanitizedEntry, ok := models.SanitizeEntry(c, &entry)
	if !ok {
		return nil
	}

	return status.Created(c, sanitizedEntry)
}

func GetEntries(c *fiber.Ctx) error {
	entries, ok := getUserEntries(c)
	if !ok {
		return nil
	}

	sanitizedEntries, ok := models.SanitizeEntries(c, &entries)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitizedEntries)
}

func GetEntry(c *fiber.Ctx) error {
	entryId, err := c.ParamsInt("entry_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid entry_id"))
	}

	entry, ok := getUserEntry(c, int64(entryId))
	if !ok {
		return nil
	}

	sanitiziedEntry, ok := models.SanitizeEntry(c, &entry)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitiziedEntry)
}

func UpdateEntry(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	entryId, err := c.ParamsInt("entry_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid entry_id"))
	}

	entry, ok := getUserEntry(c, int64(entryId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	parentFolder, err := qtx.GetFolder(*ctx, entry.FolderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return status.BadRequest(c, errors.New("invalid folder_id"))
		} else {
			return status.InternalServerError(c, nil)
		}
	}

	if parentFolder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	input, ok := schemas.GetUpdateEntryInput(c)
	if !ok {
		return nil
	}

	newEntry, err := qtx.UpdateEntry(*ctx, input)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	sanitiziedEntry, ok := models.SanitizeEntry(c, &newEntry)
	if !ok {
		return nil
	}

	return status.Ok(c, sanitiziedEntry)
}

func RemoveEntry(c *fiber.Ctx) error {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer commit()

	entryId, err := c.ParamsInt("entry_id")
	if err != nil {
		return status.BadRequest(c, errors.New("invalid entry_id"))
	}

	entry, ok := getUserEntry(c, int64(entryId))
	if !ok {
		return nil
	}

	user, ok := getUser(c)
	if !ok {
		return nil
	}

	parentFolder, err := qtx.GetFolder(*ctx, entry.FolderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return status.BadRequest(c, errors.New("invalid folder_id"))
		} else {
			return status.InternalServerError(c, nil)
		}
	}

	if parentFolder.OwnerID != user.ID {
		return status.Unauthorized(c, nil)
	}

	err = qtx.DeleteEntry(*ctx, entry.ID)
	if err != nil {
		return status.InternalServerError(c, nil)
	}

	return status.Ok(c, nil)
}

func getUserEntries(c *fiber.Ctx) ([]queries.Entry, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return []queries.Entry{}, false
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return []queries.Entry{}, false
	}

	entries, err := qtx.GetUserEntries(*ctx, user.ID)
	if err != nil {
		status.InternalServerError(c, nil)
		return []queries.Entry{}, false
	}

	return entries, true
}

func getUserEntry(c *fiber.Ctx, entryId int64) (queries.Entry, bool) {
	qtx, ctx, commit, ok := database.BeginTransaction(c)
	if !ok {
		return queries.Entry{}, false
	}
	defer commit()

	user, ok := getUser(c)
	if !ok {
		return queries.Entry{}, false
	}

	entry, err := qtx.GetUserEntry(*ctx, queries.GetUserEntryParams{
		UserID:  user.ID,
		EntryID: entryId,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status.NotFound(c, nil)
		} else {
			status.InternalServerError(c, nil)
		}

		return queries.Entry{}, false
	}

	return entry, true
}
