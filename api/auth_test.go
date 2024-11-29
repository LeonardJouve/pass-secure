package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/LeonardJouve/pass-secure/database/model"
	"github.com/LeonardJouve/pass-secure/schema"
	"github.com/LeonardJouve/pass-secure/test"
	"github.com/gofiber/fiber/v2"
)

var dummyUser = model.User{
	Email:    "email@domain.com",
	Password: "password",
}

func TestBeforeAllAuth(t *testing.T) {
	test.TestBeforeAll(t)
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

func TestAfterAllAuth(t *testing.T) {
	test.TestAfterAll(t)
}
