package handlers

import (
	"context"
	"log"
	"strings"
	"taxi-fare-calculator/database"
	"taxi-fare-calculator/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetPlaces(c *fiber.Ctx) error {
	log.Printf("📍 Fetching places from database...")
	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stations []models.Station
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("❌ Error fetching stations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching stations",
		})
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &stations); err != nil {
		log.Printf("❌ Error parsing stations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error parsing stations",
		})
	}

	// Convert stations to places format
	places := make(map[string]map[string]interface{})
	for _, station := range stations {
		// Remove "Station" suffix for display
		displayName := strings.TrimSuffix(station.Name, " Station")
		places[displayName] = map[string]interface{}{
			"stations":  []string{station.Name},
			"location":  station.Location.Coordinates,
			"connected": station.ConnectedRoutes,
		}
	}

	log.Printf("✅ Successfully fetched %d places", len(places))
	return c.JSON(fiber.Map{
		"places": places,
	})
}
