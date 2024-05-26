package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rzeradev/google-cloud-run/configs"
	"github.com/rzeradev/google-cloud-run/internal/models"
	"github.com/rzeradev/google-cloud-run/pkg/utils"
)

func FetchWeather(ctx context.Context, city string, state string) (*models.Weather, error) {
	_, span := tracer.Start(ctx, "FetchWeather")
	defer span.End()

	queryString := "country:Brazil,region:%s,name:%s"
	queryString = fmt.Sprintf(queryString, state, city)
	queryString = url.QueryEscape(queryString)
	Url := fmt.Sprintf(configs.Cfg.WeatherAPIURL, configs.Cfg.WeatherAPIKey, queryString)
	resp, err := http.Get(Url)
	if err != nil || resp.StatusCode != 200 {
		return nil, errors.New("invalid response from weatherapi")
	}
	defer resp.Body.Close()

	var apiResponse struct {
		Current struct {
			TempC float64 `json:"temp_c"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.New("failed to decode weather response")
	}

	tempC := apiResponse.Current.TempC
	weather := &models.Weather{
		TempC: tempC,
		TempF: utils.CelsiusToFahrenheit(tempC),
		TempK: utils.CelsiusToKelvin(tempC),
	}

	return weather, nil
}
