package database

import (
	"errors"
	"fmt"
	"os"

	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Database *gorm.DB

func Init(path string) error {
	var err error
	connectionURL := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("MYSQL_ROOT_PASSWORD"), os.Getenv("MYSQL_DATABASE"))
	Database, err = gorm.Open(mysql.Open(connectionURL), &gorm.Config{})

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
