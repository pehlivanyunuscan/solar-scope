package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	VictoriaMetricsURL string
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

	return &Config{
		AppPort:            appPort,
		VictoriaMetricsURL: victoriaMetricsURL,
	}
}
