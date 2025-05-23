package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/models"
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

	go func() {
		conn, release, ctx, err := database.Acquire()
		if err != nil {
			panic(err)
		}
		defer release()

		_, err = conn.Exec(ctx, "LISTEN events")
		if err != nil {
			panic(err)
		}

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Notification: %s\n", notification.Payload)
		}
	}()

	generate := func(len int) string {
		random := make([]byte, len)
		if _, err := rand.Read(random); err != nil {
			return ""
		}

		return hex.EncodeToString(random)
	}

	schemas.Init()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		qtx, ctx, commit, ok := database.BeginTransaction(c)
		if !ok {
			return nil
		}
		defer commit()

		user, err := qtx.CreateUser(ctx, queries.CreateUserParams{
			Email:    generate(10),
			Username: generate(10),
			Password: "password",
		})
		if err != nil {
			return status.BadRequest(c, err)
		}
		fmt.Println("Created user")

		user, err = qtx.UpdateUser(ctx, queries.UpdateUserParams{
			ID:       user.ID,
			Password: user.Password,
			Username: generate(10),
			Email:    generate(10),
		})
		if err != nil {
			return status.InternalServerError(c, err)
		}
		fmt.Println("Updated user")

		return status.Ok(c, models.SanitizeUser(c, &user))
	})

	app.Listen(fmt.Sprintf(":%d", PORT))
	defer app.Shutdown()

	// stop := api.Start(PORT)
	// defer stop()
}
