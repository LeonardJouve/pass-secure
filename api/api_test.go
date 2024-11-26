package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/LeonardJouve/pass-secure/database"
	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
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

const TEST_DB = "./test.db"

func TestPrepare(t *testing.T) {
	err := database.Init(TEST_DB)
	if err != nil {
		t.Fatal("Failed to initiate database connection")
	}

	schema.Init()
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

var dummyUser = model.User{
	Email:    "email@domain.com",
	Password: "password",
}

func TestRegister(t *testing.T) {
	app := fiber.New()
	app.Post("/register", Register)

	payload := schema.RegisterInput{
		Email:           dummyUser.Email,
		Password:        dummyUser.Password,
		PasswordConfirm: dummyUser.Password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	request := httptest.NewRequest("POST", "http://localhost/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal("Failed to send register")
	}

	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("Received invalid status code: expected %d, received %d", fiber.StatusCreated, response.StatusCode)
	}

	data, err := readJSONBody(response)
	if err != nil {
		t.Fatal("Failed to parse response JSON body")
	}

	_, ok := data["id"]
	if !ok {
		t.Fatal("Received invalid response")
	}
}

func TestLogin(t *testing.T) {
	app := fiber.New()
	app.Post("/login", Login)

	payload := schema.LoginInput{
		Email:    dummyUser.Email,
		Password: dummyUser.Password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	request := httptest.NewRequest("POST", "http://localhost/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal("Failed to send login")
	}

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("Received invalid status code: expected %d, received %d", fiber.StatusOK, response.StatusCode)
	}

	data, err := readJSONBody(response)
	if err != nil {
		t.Fatal("Failed to parse response JSON body")
	}

	_, ok := data["accessToken"]
	if !ok {
		t.Fatal("Received invalid response")
	}
}

func TestAfterAll(t *testing.T) {
	db, err := database.Database.DB()
	if err != nil {
		t.Fatalf("Failed to retrieve database: %v", err)
	}
	db.Close()

	err = os.Remove(TEST_DB)
	if err != nil {
		t.Fatalf("Failed to remove test database: %v", err)
	}
}
