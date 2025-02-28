package handlers

import (
	"context"
	"taxi-fare-calculator/database"
	"taxi-fare-calculator/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetStations(c *fiber.Ctx) error {
	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stations []models.Station
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching stations",
		})
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &stations); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error parsing stations",
		})
	}

	return c.JSON(fiber.Map{
		"stations": stations,
	})
}

func GetStation(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var station models.Station
	err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&station)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Station not found",
		})
	}

	return c.JSON(station)
}

func AddStation(c *fiber.Ctx) error {
	station := new(models.Station)

	if err := c.BodyParser(station); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate station data
	if station.Name == "" || len(station.Location.Coordinates) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid station data",
		})
	}

	// Set GeoJSON type if not set
	if station.Location.Type == "" {
		station.Location.Type = "Point"
	}

	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, station)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating station",
		})
	}

	station.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(fiber.StatusCreated).JSON(station)
}

func FindNearestStation(c *fiber.Ctx) error {
	lat := c.QueryFloat("lat", 0)
	lng := c.QueryFloat("lng", 0)

	if lat == 0 || lng == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid coordinates",
		})
	}

	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find nearest station using geospatial query
	pipeline := bson.A{
		bson.D{
			{Key: "$geoNear", Value: bson.D{
				{Key: "near", Value: bson.D{
					{Key: "type", Value: "Point"},
					{Key: "coordinates", Value: []float64{lng, lat}},
				}},
				{Key: "distanceField", Value: "distance"},
				{Key: "spherical", Value: true},
			}},
		},
		bson.D{{Key: "$limit", Value: 1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error finding nearest station",
		})
	}
	defer cursor.Close(ctx)

	var results []struct {
		Station  models.Station `bson:",inline"`
		Distance float64        `bson:"distance"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error parsing results",
		})
	}

	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No stations found",
		})
	}

	return c.JSON(fiber.Map{
		"station":         results[0].Station,
		"distance_meters": results[0].Distance,
	})
}

func DeleteStation(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First check if the station is referenced in any routes
	routesCollection := database.GetCollection("taxi_fare_db", "routes")
	routeCount, err := routesCollection.CountDocuments(ctx, bson.M{
		"$or": []bson.M{
			{"from": id},
			{"to": id},
		},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error checking route references",
		})
	}
	if routeCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Cannot delete station: it is referenced by existing routes",
		})
	}

	// If no routes reference this station, proceed with deletion
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error deleting station",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Station not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Station deleted successfully",
	})
}

func UpdateStation(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	station := new(models.Station)
	if err := c.BodyParser(station); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate station data
	if station.Name == "" || len(station.Location.Coordinates) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid station data",
		})
	}

	// Set GeoJSON type if not set
	if station.Location.Type == "" {
		station.Location.Type = "Point"
	}

	collection := database.GetCollection("taxi_fare_db", "stations")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"name":     station.Name,
			"image":    station.Image,
			"location": station.Location,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error updating station",
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Station not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Station updated successfully",
	})
}
