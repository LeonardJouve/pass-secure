package testing

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"gorm.io/driver/sqlite"
)

func start(t *testing.T, port uint16) func() error {
	path, err := os.Executable()
	if err != nil {
		t.Fatal("could not retrieve executable path")
	}

	databasePath := filepath.Join(filepath.Dir(path), "test.db")
	if err := database.Init(sqlite.Open(databasePath)); err != nil {
		t.Fatal(err)
	}

	schema.Init()
	model.Migrate()
	shutdown := api.Start(port)

	waitForStart()

	return shutdown
}

func waitForStart() {
	for {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthcheck", PORT))
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		time.Sleep(time.Second)
	}
}

func Test(t *testing.T) {
	var port uint16 = 3000
	shutdown := start(t, port)
	defer shutdown()

	cmd := exec.Command(fmt.Sprintf("docker run --rm -v \"$(pwd)/tests:/tests\" ovhcom/venom run /tests/test.yml --var=\"base_url=http://localhost:%d\"", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
