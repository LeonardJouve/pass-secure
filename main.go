package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/env"
	"github.com/LeonardJouve/pass-secure/schemas"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
)

const PORT = 3000

func main() {
	if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
		restore, err := env.Load(".env")
		if err != nil {
			panic(err)
		}
		defer restore()
	}

	connectionURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"))
	db, err := database.New(connectionURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Migrate()
	if err != nil {
		panic(err)
	}

	schemas.Init()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		qtx, ctx, commit, ok := database.BeginTransaction(c)
		if !ok {
			return nil
		}
		defer commit()

		usr, err := qtx.CreateUser(*ctx, queries.CreateUserParams{
			Email:    "emaillllll",
			Username: "usernameeeee",
			Password: "password",
		})
		if err != nil {
			return status.BadRequest(c, err)
		}

		user, err := qtx.GetUser(*ctx, usr.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return status.BadRequest(c, errors.New("user not found"))
			}
			return status.InternalServerError(c, nil)
		}

		return status.Ok(c, fiber.Map{
			"email":    user.Email,
			"username": user.Username,
			"password": user.Password,
		})
	})

	app.Listen(fmt.Sprintf(":%d", PORT))
	defer app.Shutdown()

	// stop := api.Start(PORT)
	// defer stop()
}
