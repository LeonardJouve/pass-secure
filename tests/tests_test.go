package testing

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/LeonardJouve/pass-secure/api"
	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/env"
	"github.com/LeonardJouve/pass-secure/schemas"
)

func start(t *testing.T, port uint16) func() error {
	if os.Getenv("ENVIRONMENT") != "PRODUCTION" {
		restore, err := env.Load(".env")
		if err != nil {
			t.Fatal(err)
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
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Migrate()
	if err != nil {
		t.Fatal(err)
	}

	schemas.Init()

	shutdown, err := api.Start(port)
	if err != nil {
		t.Fatal(err)
	}

	waitForStart(port)

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
	portString := os.Getenv("PORT")
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		t.Fatal(err)
	}

	shutdown := start(t, uint16(port))
	defer shutdown()

	cmd := exec.Command(fmt.Sprintf("docker run --rm -v \"$(pwd)/tests:/tests\" ovhcom/venom run /tests/test.yml --var=\"base_url=http://localhost:%d\"", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
