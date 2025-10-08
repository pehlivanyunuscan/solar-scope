package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
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

	port := "8080"
	log.Printf("Starting server on port %s", port)

	err := app.Listen("0.0.0.0:" + port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
