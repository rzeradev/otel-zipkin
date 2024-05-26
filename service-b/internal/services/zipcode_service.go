package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rzeradev/google-cloud-run/configs"
	"go.opentelemetry.io/otel"
)

type Location struct {
	City  string `json:"localidade"`
	State string `json:"uf"`
}

var brazilStates = map[string]string{
	"AC": "Acre",
	"AL": "Alagoas",
	"AP": "Amapá",
	"AM": "Amazonas",
	"BA": "Bahia",
	"CE": "Ceará",
	"DF": "Distrito Federal",
	"ES": "Espírito Santo",
	"GO": "Goiás",
	"MA": "Maranhão",
	"MT": "Mato Grosso",
	"MS": "Mato Grosso do Sul",
	"MG": "Minas Gerais",
	"PA": "Pará",
	"PB": "Paraíba",
	"PR": "Paraná",
	"PE": "Pernambuco",
	"PI": "Piauí",
	"RJ": "Rio de Janeiro",
	"RN": "Rio Grande do Norte",
	"RS": "Rio Grande do Sul",
	"RO": "Rondônia",
	"RR": "Roraima",
	"SC": "Santa Catarina",
	"SP": "São Paulo",
	"SE": "Sergipe",
	"TO": "Tocantins",
}

var tracer = otel.Tracer("orchestrator-service")

func FetchLocation(ctx context.Context, zipcode string) (*Location, error) {
	_, span := tracer.Start(ctx, "FetchLocation")
	defer span.End()

	url := fmt.Sprintf(configs.Cfg.CepAPIURL, zipcode)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, errors.New("invalid response from viacep")
	}
	defer resp.Body.Close()

	var location Location
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return nil, errors.New("failed to decode location response")
	}

	if location.City == "" {
		return nil, errors.New("city not found for the given zipcode")
	}

	// Check if location.State matches a key in brazilStates and swap the value
	if fullStateName, exists := brazilStates[location.State]; exists {
		location.State = fullStateName
	}

	return &location, nil
}
