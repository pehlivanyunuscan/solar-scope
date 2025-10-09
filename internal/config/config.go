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
	DBHost             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBPort             string
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

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // Varsayılan DB host
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "solar_scope_user" // Varsayılan DB user
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "1" // Varsayılan DB password
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "solar_scope_db" // Varsayılan DB name
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Varsayılan DB port
	}

	return &Config{
		AppPort:            appPort,
		VictoriaMetricsURL: victoriaMetricsURL,
		SolarForecasterURL: solarForecasterURL,
		DBHost:             dbHost,
		DBUser:             dbUser,
		DBPassword:         dbPassword,
		DBName:             dbName,
		DBPort:             dbPort,
	}
}
