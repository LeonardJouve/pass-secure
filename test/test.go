package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
)

func getTestDatabasePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "..", "database.db"), nil
}

func TestBeforeAll(t *testing.T) {
	databasePath, err := getTestDatabasePath()
	if err != nil {
		t.Fatal("Failed to retrieve database path")
	}

	err = database.Init(databasePath)
	if err != nil {
		t.Fatal("Failed to initiate database connection")
	}

	schema.Init()
	model.Migrate()
}

func TestAfterAll(t *testing.T) {
	db, err := database.Database.DB()
	if err != nil {
		t.Fatalf("Failed to retrieve database: %v", err)
	}
	db.Close()

	databasePath, err := getTestDatabasePath()
	if err != nil {
		t.Fatal("Failed to retrieve database path")
	}

	err = os.Remove(databasePath)
	if err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}
