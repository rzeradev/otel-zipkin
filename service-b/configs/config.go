package configs

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	WeatherAPIKey string `mapstructure:"WEATHER_API_KEY"`
	CepAPIURL     string `mapstructure:"CEP_API_URL"`
	WeatherAPIURL string `mapstructure:"WEATHER_API_URL"`
	ServerPort    string `mapstructure:"SERVER_PORT"`
}

var Cfg *Config

func defaultAndBindings() error {
	defaultConfigs := map[string]string{
		"SERVER_PORT":     "8181",
		"WEATHER_API_KEY": "",
		"WEATHER_API_URL": "https://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		"CEP_API_URL":     "https://viacep.com.br/ws/%s/json/",
	}
	for envKey, envValue := range defaultConfigs {
		err := viper.BindEnv(envKey)
		if err != nil {
			return err
		}
		viper.SetDefault(envKey, envValue)
	}
	return nil

}
func LoadConfig(workdir string) (*Config, error) {
	viper.SetConfigName("app_config")

	envFilePath := path.Join(workdir, ".env")
	if _, err := os.Stat(envFilePath); err == nil {
		viper.SetConfigType("env")
		viper.AddConfigPath(workdir)
		viper.SetConfigFile(envFilePath)

		err = viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("error reading .env file: %w", err)
		}

		// fmt.Println("Reading .env file")
		// fmt.Println(workdir)
	} else if os.IsNotExist(err) {
		fmt.Println(".env file does not exist, skipping")
	} else {
		return nil, fmt.Errorf("error checking .env file: %w", err)
	}

	viper.AutomaticEnv()

	err := defaultAndBindings()
	if err != nil {
		return nil, fmt.Errorf("error in defaultAndBindings: %w", err)
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return Cfg, nil
}
