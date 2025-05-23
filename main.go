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

	// _, err = db.Exec("LISTEN events")
	// if err != nil {
	// 	panic(err)
	// }

	// go func() {
	// 	for {
	// 		notification, err := db.WaitForNotification()
	// 		if err != nil {
	// 			fmt.Printf("test %s", err.Error())
	// 			panic(err)
	// 		}

	// 		fmt.Printf("Notification: %s\n", notification.Payload)
	// 	}
	// }()

	schemas.Init()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		qtx, ctx, commit, ok := database.BeginTransaction(c)
		if !ok {
			return nil
		}
		defer commit()

		users, err := qtx.GetUsers(*ctx)
		if err != nil {
			return status.BadRequest(c, err)
		}
		for _, user := range users {
			fmt.Println(user.Email)
		}
		fmt.Println("Selected users")

		usr, err := qtx.CreateUser(*ctx, queries.CreateUserParams{
			Email:    "cdasda",
			Username: "dwadaw",
			Password: "password",
		})
		if err != nil {
			return status.BadRequest(c, err)
		}
		fmt.Println("Created user")

		user, err := qtx.GetUser(*ctx, usr.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return status.BadRequest(c, errors.New("user not found"))
			}
			return status.InternalServerError(c, err)
		}
		fmt.Println("Selected user")

		user, err = qtx.UpdateUser(*ctx, queries.UpdateUserParams{
			Email: "testttttt",
		})
		if err != nil {
			return status.InternalServerError(c, err)
		}
		fmt.Println("Updated user")

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
