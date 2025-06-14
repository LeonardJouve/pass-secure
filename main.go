package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/env"
	"github.com/LeonardJouve/pass-secure/schemas"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
		restore, err := env.Load(".env")
		if err != nil {
			panic(err)
		}
		defer restore()
	}

	connectionURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"))
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

	portString := os.Getenv("PORT")
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		panic(err)
	}

	stop, err := api.Start(uint16(port))
	if err != nil {
		panic(err)
	}
	defer stop()
}
