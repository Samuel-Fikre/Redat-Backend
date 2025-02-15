package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"taxi-fare-calculator/config"
	"taxi-fare-calculator/database"
	"taxi-fare-calculator/handlers"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load configuration
	log.Printf("🚀 Starting Taxi Fare Calculator API...")
	cfg := config.LoadConfig()

	// Connect to database with retries
	maxRetries := 3
	var err error
	for i := 0; i < maxRetries; i++ {
		err = database.ConnectDB(cfg.MongoURI)
		if err == nil {
			break
		}
		log.Printf("❌ Failed to connect to MongoDB (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("❌ Failed to initialize database after %d attempts: %v", maxRetries, err)
	}
	defer database.DisconnectDB()

	// Initialize Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Station Routes
	app.Get("/stations", handlers.GetStations)
	app.Get("/stations/:id", handlers.GetStation)
	app.Post("/stations", handlers.AddStation)
	app.Delete("/stations/:id", handlers.DeleteStation)
	app.Put("/stations/:id", handlers.UpdateStation)

	// Route Routes
	app.Get("/routes", handlers.GetRoutes)
	app.Get("/route", handlers.GetRoute)
	app.Post("/routes", handlers.AddRoute)
	app.Put("/routes/:id", handlers.UpdateRoute)
	app.Delete("/routes/:id", handlers.DeleteRoute)
	app.Get("/journey", handlers.CalculateJourney)
	app.Get("/nearest-station", handlers.FindNearestStation)
	app.Get("/route-map", handlers.GetRouteWithMap)
	app.Get("/places", handlers.GetPlaces)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": "connected",
		})
	})

	// Add static file serving
	app.Static("/static", "./static")
	app.Get("/", handlers.ServeMapUI)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("🛑 Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	log.Printf("🌍 Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
