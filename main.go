package main

import (
	"os"
	"path/filepath"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
)

func main() {
	schema.Init()

	path, err := os.Executable()
	if err != nil {
		panic("Could not retrieve executable path")
	}
	databasePath := filepath.Join(filepath.Dir(path), "database.db")
	err = database.Init(databasePath)
	if err != nil {
		panic("Could not initialize database")
	}

	model.Migrate()

	shutdown := api.Start(3000)
	defer shutdown()
}
