package main

import (
	"fmt"
	"log"
	"os"
	"solar-scope/database"
	"solar-scope/internal/client"
	"solar-scope/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Konfigürasyonu yükle
	cfg := config.LoadConfig()
	log.Printf("Config loaded: %+v", cfg)
	database.Connect(*cfg)

	vmClient, err := client.NewPrometheusClient(cfg.VictoriaMetricsURL)
	if err != nil {
		log.Fatalf("Error creating VictoriaMetrics client: %v", err)
	}
	log.Println("VictoriaMetrics client created successfully:", vmClient)

	sfClient := client.NewSolarForecasterClient(cfg.SolarForecasterURL)

	log.Println("SolarForecaster client created successfully:", sfClient)
	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New()) // Panik durumlarında uygulamanın çökmesini önler

	//API rotalarını gruplayalım
	apiV1 := app.Group("/api/v1")

	apiV1.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "API is healthy",
		})
	})

	apiV1.Get("/panel/metrics", func(c *fiber.Ctx) error {
		query := `mppt_values{sensor="panel gucu"}`

		result, err := vmClient.Query(query)
		if err != nil {
			log.Printf("Error querying VictoriaMetrics: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to query VictoriaMetrics",
			})
		}

		return c.Status(fiber.StatusOK).JSON(result)
	})

	//ML API rotaları
	forecasterGroup := apiV1.Group("/forecaster")

	//JSON ile anlık tahmin isteği
	forecasterGroup.Post("/run", func(c *fiber.Ctx) error {
		// İstek gövdesini oku
		var reqPayload client.RunRequest
		if err := c.BodyParser(&reqPayload); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid request payload",
			})
		}
		result, err := sfClient.RunForecast(reqPayload)
		if err != nil {
			log.Printf("Error calling RunForecast: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to run forecast",
			})
		}

		go database.SaveResultToDB(result)

		return c.Status(200).JSON(result)
	})
	// .env dosyası yükle
	forecasterGroup.Post("/upload-env", func(c *fiber.Ctx) error {
		// Dosyayı oku
		file, err := c.FormFile("env_file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to read env file",
			})
		}

		// Geçici bir dosyaya kaydet
		tempPath := fmt.Sprintf("./temp_%s", file.Filename)
		if err := c.SaveFile(file, tempPath); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to save env file",
			})
		}
		defer os.Remove(tempPath) // İşlem sonrası dosyayı sil
		result, err := sfClient.UploadEnvFile(tempPath)
		if err != nil {
			log.Printf("Error calling UploadEnvFile: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to upload env file",
			})
		}
		return c.Status(200).JSON(result)
	})
	// session_id ile tahmin isteği (opsiyonel overrides ile)
	forecasterGroup.Post("/run-with-env/:session_id", func(c *fiber.Ctx) error {
		sessionID := c.Params("session_id")
		var overrides map[string]interface{}
		// Body boş değilse, overrides'ı ayrıştır
		if len(c.Body()) > 0 {
			if err := c.BodyParser(&overrides); err != nil {
				return c.Status(400).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid overrides payload",
				})
			}
		}
		result, err := sfClient.RunWithEnv(sessionID, overrides)
		if err != nil {
			log.Printf("Error calling RunWithEnv: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to run with env",
			})
		}

		go database.SaveResultToDB(result)

		return c.Status(200).JSON(result)
	})
	// Mevcut session'ları listele
	forecasterGroup.Get("/sessions", func(c *fiber.Ctx) error {
		result, err := sfClient.GetSessions()
		if err != nil {
			log.Printf("Error calling GetSessions: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to get sessions",
			})
		}
		return c.Status(200).JSON(result)
	})

	// session sil
	forecasterGroup.Delete("/sessions/:session_id", func(c *fiber.Ctx) error {
		sessionID := c.Params("session_id")
		result, err := sfClient.DeleteSession(sessionID)
		if err != nil {
			log.Printf("Error calling DeleteSession: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to delete session",
			})
		}
		return c.Status(200).JSON(result)
	})

	// Örnek .env dosyasını al
	forecasterGroup.Get("/sample-env", func(c *fiber.Ctx) error {
		result, err := sfClient.GetSampleEnv()
		if err != nil {
			log.Printf("Error calling GetSampleEnv: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to get sample env",
			})
		}
		return c.Status(200).JSON(result)
	})

	forecastsGroup := apiV1.Group("/forecasts")
	// Depolanan tahminleri listele
	forecastsGroup.Get("/", func(c *fiber.Ctx) error {
		forecasts, err := database.GetRecentForecasts(10) // Son 10 tahmini al
		if err != nil {
			log.Printf("Error retrieving forecasts: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve forecasts",
			})
		}
		return c.Status(200).JSON(forecasts)
	})

	// Belirli bir tahmini ID ile al
	forecastsGroup.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		forecast, err := database.GetForecastByID(id)
		if err != nil {
			log.Printf("Error retrieving forecast by ID: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve forecast",
			})
		}
		if forecast == nil {
			return c.Status(404).JSON(fiber.Map{
				"status":  "error",
				"message": "Forecast not found",
			})
		}
		return c.Status(200).JSON(forecast)
	})

	log.Printf("Starting server on port %s", cfg.AppPort)

	err = app.Listen("0.0.0.0:" + cfg.AppPort)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
