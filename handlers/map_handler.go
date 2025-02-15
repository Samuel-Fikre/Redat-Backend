package handlers

import (
	"taxi-fare-calculator/database"
	"taxi-fare-calculator/models"
	"taxi-fare-calculator/utils"

	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type RouteResponse struct {
	Route      []models.Station  `json:"route"`
	Path       interface{}       `json:"path"`
	TotalPrice float64           `json:"total_price"`
	Distance   float64           `json:"distance"`
	Duration   float64           `json:"duration"`
	Legs       []models.RouteLeg `json:"legs"`
}

func GetRouteWithMap(c *fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	userLat := c.QueryFloat("user_lat", 0)
	userLng := c.QueryFloat("user_lng", 0)

	if from == "" || to == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Both 'from' and 'to' parameters are required",
		})
	}

	// Convert place names to station names only if they don't already end with "Station"
	fromStation := from
	toStation := to
	if !strings.HasSuffix(from, "Station") {
		fromStation = from + " Station"
	}
	if !strings.HasSuffix(to, "Station") {
		toStation = to + " Station"
	}

	// Log the actual station names being used
	log.Printf("Converting route from %s to %s", fromStation, toStation)

	collection := database.GetCollection("taxi_fare_db", "routes")
	journey, err := models.CalculateJourney(fromStation, toStation, collection)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Initialize map service
	mapService := utils.NewMapService()

	var completeRoute RouteResponse
	completeRoute.Route = journey.Stations
	completeRoute.TotalPrice = journey.TotalPrice
	completeRoute.Legs = journey.Legs

	// If user location is provided, get route to first station
	if userLat != 0 && userLng != 0 {
		firstStation := journey.Stations[0]
		walkingRoute, err := mapService.GetRoute(
			userLng, userLat,
			firstStation.Location.Coordinates[0],
			firstStation.Location.Coordinates[1],
		)
		if err == nil {
			completeRoute.Path = walkingRoute.Routes[0].Geometry
			completeRoute.Distance = walkingRoute.Routes[0].Distance
			completeRoute.Duration = walkingRoute.Routes[0].Duration
		}
	}

	// Get driving route between stations
	var totalDistance, totalDuration float64
	for i := 0; i < len(journey.Stations)-1; i++ {
		from := journey.Stations[i]
		to := journey.Stations[i+1]

		route, err := mapService.GetRoute(
			from.Location.Coordinates[0], from.Location.Coordinates[1],
			to.Location.Coordinates[0], to.Location.Coordinates[1],
		)
		if err == nil {
			totalDistance += route.Routes[0].Distance
			totalDuration += route.Routes[0].Duration
		}
	}

	completeRoute.Distance += totalDistance
	completeRoute.Duration += totalDuration

	return c.JSON(completeRoute)
}
