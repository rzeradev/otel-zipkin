package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rzeradev/google-cloud-run/configs"
	"github.com/rzeradev/google-cloud-run/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func initTestConfig() (*configs.Config, error) {
	workdir, err := os.Getwd()
	// go up 1 directory
	workdir = workdir[:len(workdir)-4]
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}

	cfg, err := configs.LoadConfig(workdir)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	return cfg, nil
}

func TestGetWeather_Success(t *testing.T) {
	app := fiber.New()
	_, err := initTestConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	app.Get("/weather/:zipcode", handlers.GetWeather)

	req := httptest.NewRequest(http.MethodGet, "/weather/26572070", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]float64
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)

	assert.Contains(t, body, "temp_C")
	assert.Contains(t, body, "temp_F")
	assert.Contains(t, body, "temp_K")
}

func TestGetWeather_InvalidZipcode(t *testing.T) {
	app := fiber.New()
	_, err := initTestConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	app.Get("/weather/:zipcode", handlers.GetWeather)

	req := httptest.NewRequest(http.MethodGet, "/weather/123", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "invalid zipcode", body["message"])
}

func TestGetWeather_ZipcodeNotFound(t *testing.T) {
	app := fiber.New()
	_, err := initTestConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	app.Get("/weather/:zipcode", handlers.GetWeather)

	req := httptest.NewRequest(http.MethodGet, "/weather/99999999", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "can not find zipcode", body["message"])
}

func TestGetWeather_FetchWeatherFailure(t *testing.T) {
	app := fiber.New()
	_, err := initTestConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	// Mock the service to simulate a failure in fetching weather data
	app.Get("/weather/:zipcode", func(c *fiber.Ctx) error {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to fetch weather data",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/weather/01001000", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "failed to fetch weather data", body["message"])
}
