package handlers

import (
	"context"
	"strings"
	"taxi-fare-calculator/database"
	"taxi-fare-calculator/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetRoutes(c *fiber.Ctx) error {
	collection := database.GetCollection("taxi_fare_db", "routes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var routes []models.Route
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching routes",
		})
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &routes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error parsing routes",
		})
	}

	return c.JSON(fiber.Map{
		"routes": routes,
	})
}

func GetRoute(c *fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing from or to parameters",
		})
	}

	// Get current time in Addis Ababa
	location, err := time.LoadLocation("Africa/Addis_Ababa")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load timezone",
		})
	}
	now := time.Now().In(location)
	hour := now.Hour()
	minute := now.Minute()
	currentTime := float64(hour) + float64(minute)/60.0

	// Check if it's night time (18:30 - 22:30)
	isNight := currentTime >= 18.5 && currentTime <= 22.5

	collection := database.GetCollection("taxi_fare_db", "routes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First try to find a direct route
	var route models.Route
	err = collection.FindOne(ctx, bson.M{
		"from": from,
		"to":   to,
	}).Decode(&route)

	if err == nil {
		// Found a direct route
		// Apply night fare if applicable
		price := route.Price
		if isNight {
			price = price * 1.4 // 40% increase for night fare
		}
		return c.JSON(models.JourneyResponse{
			Route:      []string{from, to},
			TotalPrice: price,
			Legs: []models.RouteLeg{{
				From:  from,
				To:    to,
				Price: price,
			}},
			IsNight: isNight,
		})
	}

	// If no direct route, find all possible routes
	var routes []models.Route
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error searching for routes",
		})
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &routes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error parsing routes",
		})
	}

	// Find the best path using the routes
	path, totalPrice, legs := findBestPath(routes, from, to)
	if len(path) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No route found",
		})
	}

	// Apply night fare if applicable
	if isNight {
		totalPrice = totalPrice * 1.4 // 40% increase for night fare
		for i := range legs {
			legs[i].Price = legs[i].Price * 1.4
		}
	}

	return c.JSON(models.JourneyResponse{
		Route:      path,
		TotalPrice: totalPrice,
		Legs:       legs,
		IsNight:    isNight,
	})
}

// findBestPath finds the shortest path between two stations using available routes
func findBestPath(routes []models.Route, from, to string) ([]string, float64, []models.RouteLeg) {
	// Create a graph representation
	graph := make(map[string]map[string]float64)
	for _, route := range routes {
		if graph[route.From] == nil {
			graph[route.From] = make(map[string]float64)
		}
		graph[route.From][route.To] = route.Price

		// Add reverse direction if it doesn't exist
		if graph[route.To] == nil {
			graph[route.To] = make(map[string]float64)
		}
		if _, exists := graph[route.To][route.From]; !exists {
			graph[route.To][route.From] = route.Price
		}

		// Add connections through intermediate stations
		if !route.IsDirectRoute && len(route.IntermediateStations) > 0 {
			prevStation := route.From
			totalPrice := route.Price / float64(len(route.IntermediateStations)+1)

			for _, station := range route.IntermediateStations {
				if graph[prevStation] == nil {
					graph[prevStation] = make(map[string]float64)
				}
				graph[prevStation][station] = totalPrice

				if graph[station] == nil {
					graph[station] = make(map[string]float64)
				}
				graph[station][prevStation] = totalPrice

				prevStation = station
			}

			// Connect last intermediate station to destination
			if graph[prevStation] == nil {
				graph[prevStation] = make(map[string]float64)
			}
			graph[prevStation][route.To] = totalPrice

			if graph[route.To] == nil {
				graph[route.To] = make(map[string]float64)
			}
			graph[route.To][prevStation] = totalPrice
		}
	}

	// Use Dijkstra's algorithm to find the shortest path
	distances := make(map[string]float64)
	previous := make(map[string]string)
	unvisited := make(map[string]bool)

	// Initialize distances
	for station := range graph {
		distances[station] = float64(^uint(0) >> 1) // Max float64
		unvisited[station] = true
	}
	distances[from] = 0

	for len(unvisited) > 0 {
		// Find unvisited node with minimum distance
		var current string
		minDist := float64(^uint(0) >> 1)
		for station := range unvisited {
			if distances[station] < minDist {
				current = station
				minDist = distances[station]
			}
		}

		if current == "" || current == to {
			break
		}

		delete(unvisited, current)

		// Update distances to neighbors
		for neighbor, price := range graph[current] {
			if !unvisited[neighbor] {
				continue
			}
			newDist := distances[current] + price
			if newDist < distances[neighbor] {
				distances[neighbor] = newDist
				previous[neighbor] = current
			}
		}
	}

	// Reconstruct path
	if distances[to] == float64(^uint(0)>>1) {
		return nil, 0, nil
	}

	path := []string{to}
	current := to
	legs := []models.RouteLeg{}
	for current != from {
		prev := previous[current]
		path = append([]string{prev}, path...)
		legs = append([]models.RouteLeg{{
			From:  prev,
			To:    current,
			Price: graph[prev][current],
		}}, legs...)
		current = prev
	}

	return path, distances[to], legs
}

func AddRoute(c *fiber.Ctx) error {
	route := new(models.Route)
	if err := c.BodyParser(route); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate route data
	if route.From == "" || route.To == "" || route.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid route data",
		})
	}

	// Validate intermediate stations if not a direct route
	if !route.IsDirectRoute {
		if len(route.IntermediateStations) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Non-direct route must have intermediate stations",
			})
		}
		// Check for duplicate stations
		stations := make(map[string]bool)
		stations[route.From] = true
		for _, station := range route.IntermediateStations {
			if station == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid intermediate station",
				})
			}
			if stations[station] {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Duplicate stations in route",
				})
			}
			stations[station] = true
		}
		if stations[route.To] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Duplicate stations in route",
			})
		}
	} else {
		route.IntermediateStations = nil // Ensure no intermediate stations for direct routes
	}

	collection := database.GetCollection("taxi_fare_db", "routes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if route already exists
	existingRoute := collection.FindOne(ctx, bson.M{
		"from": route.From,
		"to":   route.To,
	})
	if existingRoute.Err() == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Route already exists",
		})
	}

	result, err := collection.InsertOne(ctx, route)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating route",
		})
	}

	route.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(fiber.StatusCreated).JSON(route)
}

func UpdateRoute(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	route := new(models.Route)
	if err := c.BodyParser(route); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Helper function to ensure consistent station names
	getStationName := func(name string) string {
		name = strings.TrimSuffix(name, " Station")
		return name
	}

	// Clean up station names for consistency
	route.From = getStationName(route.From)
	route.To = getStationName(route.To)
	for i, station := range route.IntermediateStations {
		route.IntermediateStations[i] = getStationName(station)
	}

	// Validate route data
	if route.From == "" || route.To == "" || route.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid route data",
		})
	}

	// Validate intermediate stations if not a direct route
	if !route.IsDirectRoute {
		if len(route.IntermediateStations) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Non-direct route must have intermediate stations",
			})
		}
		// Check for duplicate stations
		stations := make(map[string]bool)
		stations[route.From] = true
		for _, station := range route.IntermediateStations {
			if station == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid intermediate station",
				})
			}
			if stations[station] {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Duplicate stations in route",
				})
			}
			stations[station] = true
		}
		if stations[route.To] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Duplicate stations in route",
			})
		}
	} else {
		route.IntermediateStations = nil // Ensure no intermediate stations for direct routes
	}

	collection := database.GetCollection("taxi_fare_db", "routes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"from":                 route.From,
			"to":                   route.To,
			"price":                route.Price,
			"isDirectRoute":        route.IsDirectRoute,
			"intermediateStations": route.IntermediateStations,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error updating route",
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Route updated successfully",
	})
}

func DeleteRoute(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	collection := database.GetCollection("taxi_fare_db", "routes")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error deleting route",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Route deleted successfully",
	})
}

func CalculateJourney(c *fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Both 'from' and 'to' parameters are required",
		})
	}

	collection := database.GetCollection("taxi_fare_db", "routes")
	journey, err := models.CalculateJourney(from, to, collection)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(journey)
}
