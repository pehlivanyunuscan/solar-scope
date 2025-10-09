package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	VictoriaMetricsURL string
	SolarForecasterURL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080" // Varsayılan port
	}

	victoriaMetricsURL := os.Getenv("VICTORIA_METRICS_URL")
	if victoriaMetricsURL == "" {
		victoriaMetricsURL = "http://localhost:8428" // Varsayılan VictoriaMetrics URL'si
	}

	// SolarForecaster URL'sini al
	solarForecasterURL := os.Getenv("SOLAR_FORECASTER_URL")
	if solarForecasterURL == "" {
		solarForecasterURL = "http://10.67.67.192:4545" // Varsayılan SolarForecaster URL'si
	}

	return &Config{
		AppPort:            appPort,
		VictoriaMetricsURL: victoriaMetricsURL,
		SolarForecasterURL: solarForecasterURL,
	}
}
