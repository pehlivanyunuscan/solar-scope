package main

import (
	"log"
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

	promClient, err := client.NewPrometheusClient(cfg.PrometheusURL)
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v", err)
	}
	log.Println("Prometheus client created successfully:", promClient)

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

		result, err := promClient.Query(query)
		if err != nil {
			log.Printf("Error querying Prometheus: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to query Prometheus",
			})
		}

		return c.Status(fiber.StatusOK).JSON(result)
	})

	log.Printf("Starting server on port %s", cfg.AppPort)

	err = app.Listen("0.0.0.0:" + cfg.AppPort)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
