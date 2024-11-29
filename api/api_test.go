package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func readJSONBody(response *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func TestHealthCare(t *testing.T) {
	app := fiber.New()
	app.Get("/healthcheck", HealthCheck)

	request := httptest.NewRequest("GET", "http://localhost/healthcheck", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal("Failed to send ping")
	}

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("Received invalid status code: expected %d, received %d", fiber.StatusOK, response.StatusCode)
	}

	data, err := readJSONBody(response)
	if err != nil {
		t.Fatal("Failed to parse response JSON body")
	}

	if data["status"] != "ok" {
		t.Fatalf("Received invalid status: expected ok, received %s", data["status"])
	}
}
