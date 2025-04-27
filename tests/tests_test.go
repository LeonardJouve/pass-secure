package testing

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/env"
	"github.com/LeonardJouve/pass-secure/schemas"
)

const PORT = 3000

func start(t *testing.T) func() error {
	if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
		restore, err := env.Load(".env")
		if err != nil {
			t.Fatal(err)
		}
		defer restore()
	}

	connectionURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME"))
	db, err := database.New(connectionURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Migrate()
	if err != nil {
		panic(err)
	}

	schemas.Init()
	shutdown := api.Start(PORT)

	waitForStart(PORT)

	return shutdown
}

func waitForStart(port uint16) {
	for {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthcheck", port))
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
	shutdown := start(t)
	defer shutdown()

	cmd := exec.Command(fmt.Sprintf("docker run --rm -v \"$(pwd)/tests:/tests\" ovhcom/venom run /tests/test.yml --var=\"base_url=http://localhost:%d\"", PORT))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
