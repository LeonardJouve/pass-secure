package api

import (
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

func GetEntries(c *fiber.Ctx) error {
	tx, ok := database.BeginTransaction(c)
	if !ok {
		return nil
	}
	defer database.CommitTransactionIfSuccess(c, tx)

	// tx.Model(&Model.Entr)

	return status.Ok(c, nil)
}
