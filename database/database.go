package database

import (
	"errors"

	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Database *gorm.DB

func Init(path string) error {
	var err error
	Database, err = gorm.Open(sqlite.Open(path), &gorm.Config{})

	return err
}

func CommitTransactionIfSuccess(c *fiber.Ctx, tx *gorm.DB) {
	if c.Response().StatusCode()/100 != 2 {
		tx.Rollback()
	}

	tx.Commit()
}

func BeginTransaction(c *fiber.Ctx) (*gorm.DB, bool) {
	tx := Database.Begin()
	if tx.Error != nil {
		status.InternalServerError(c, nil)
		return nil, false
	}

	return tx, true
}

func Execute(c *fiber.Ctx, err error) bool {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		status.InternalServerError(c, err)
		return false
	}

	return true
}
