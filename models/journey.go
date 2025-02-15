package models

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"taxi-fare-calculator/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Journey struct {
	Stations   []Station  `json:"stations"`
	TotalPrice float64    `json:"total_price"`
	Legs       []RouteLeg `json:"legs"`
}

// calculateDistance calculates the distance between two points using the Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func CalculateJourney(from, to string, collection *mongo.Collection) (*Journey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Calculating journey from %s to %s", from, to)

	// Helper function to ensure consistent station names
	getStationName := func(name string) string {
		name = strings.TrimSuffix(name, " Station")
		return name + " Station"
	}

	// Normalize station names
	from = getStationName(from)
	to = getStationName(to)

	// Check if source and destination are the same
	if from == to {
		// Get station details
		stationsColl := collection.Database().Collection("stations")
		station := Station{}
		err := stationsColl.FindOne(ctx, bson.M{"name": from}).Decode(&station)
		if err != nil {
			return nil, fmt.Errorf("station not found: %v", err)
		}

		// Return a journey with zero price and same station
		return &Journey{
			Stations:   []Station{station},
			TotalPrice: 0,
			Legs:       []RouteLeg{},
		}, nil
	}

	// Find the route
	var route Route
	err := collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"from": from, "to": to},
			{"to": from, "from": to},
		},
	}).Decode(&route)

	if err != nil {
		// If no route found, calculate based on distance
		stationsColl := collection.Database().Collection("stations")
		fromStation := Station{}
		toStation := Station{}

		err1 := stationsColl.FindOne(ctx, bson.M{"name": from}).Decode(&fromStation)
		err2 := stationsColl.FindOne(ctx, bson.M{"name": to}).Decode(&toStation)

		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("error fetching station details")
		}

		// Calculate distance between stations
		distance := calculateDistance(
			fromStation.Location.Coordinates[1],
			fromStation.Location.Coordinates[0],
			toStation.Location.Coordinates[1],
			toStation.Location.Coordinates[0],
		)

		// Calculate fare based on distance
		fare := utils.CalculateFare(distance)

		return &Journey{
			Stations:   []Station{fromStation, toStation},
			TotalPrice: fare,
			Legs: []RouteLeg{
				{
					From:  from,
					To:    to,
					Price: fare,
				},
			},
		}, nil
	}

	stationsColl := collection.Database().Collection("stations")

	// If it's a direct route, return journey with just start and end stations
	if route.IsDirectRoute {
		fromStation := Station{}
		toStation := Station{}

		err1 := stationsColl.FindOne(ctx, bson.M{"name": from}).Decode(&fromStation)
		err2 := stationsColl.FindOne(ctx, bson.M{"name": to}).Decode(&toStation)

		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("error fetching station details")
		}

		return &Journey{
			Stations:   []Station{fromStation, toStation},
			TotalPrice: route.Price,
			Legs: []RouteLeg{
				{
					From:  from,
					To:    to,
					Price: route.Price,
				},
			},
		}, nil
	}

	// For routes with intermediate stations
	if len(route.IntermediateStations) > 0 {
		// Get all station details
		fromStation := Station{}
		toStation := Station{}
		var intermediateStations []Station

		err1 := stationsColl.FindOne(ctx, bson.M{"name": from}).Decode(&fromStation)
		err2 := stationsColl.FindOne(ctx, bson.M{"name": to}).Decode(&toStation)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("error fetching station details")
		}

		// Get intermediate station details
		for _, intStationName := range route.IntermediateStations {
			intStationName = getStationName(intStationName)
			var intStation Station
			err := stationsColl.FindOne(ctx, bson.M{"name": intStationName}).Decode(&intStation)
			if err != nil {
				return nil, fmt.Errorf("error fetching intermediate station details")
			}
			intermediateStations = append(intermediateStations, intStation)
		}

		// Build the complete stations list
		stations := []Station{fromStation}
		stations = append(stations, intermediateStations...)
		stations = append(stations, toStation)

		// Calculate prices for each segment
		var legs []RouteLeg
		var totalKnownPrice float64
		var unknownSegments int

		// First pass: Calculate known prices and count unknown segments
		for i := 0; i < len(stations)-1; i++ {
			currentStation := stations[i]
			nextStation := stations[i+1]

			log.Printf("Looking for route between %s and %s", currentStation.Name, nextStation.Name)

			// Try to find existing route price
			var segmentRoute Route

			// Get base names without "Station" suffix
			currentName := strings.TrimSuffix(currentStation.Name, " Station")
			nextName := strings.TrimSuffix(nextStation.Name, " Station")

			// Try all possible combinations of names (with and without "Station" suffix)
			filter := bson.M{
				"$or": []bson.M{
					// With "Station" suffix
					{
						"from":          currentStation.Name,
						"to":            nextStation.Name,
						"isDirectRoute": true,
					},
					{
						"from":          nextStation.Name,
						"to":            currentStation.Name,
						"isDirectRoute": true,
					},
					// Without "Station" suffix
					{
						"from":          currentName,
						"to":            nextName,
						"isDirectRoute": true,
					},
					{
						"from":          nextName,
						"to":            currentName,
						"isDirectRoute": true,
					},
				},
			}

			log.Printf("Filter: %+v", filter)

			err := collection.FindOne(ctx, filter).Decode(&segmentRoute)
			if err == nil {
				log.Printf("Found existing route price: %f", segmentRoute.Price)
				// Found existing route price
				legs = append(legs, RouteLeg{
					From:  currentStation.Name,
					To:    nextStation.Name,
					Price: segmentRoute.Price,
				})
				totalKnownPrice += segmentRoute.Price
			} else {
				log.Printf("No existing route found: %v", err)
				// Price unknown for this segment
				unknownSegments++
				legs = append(legs, RouteLeg{
					From:  currentStation.Name,
					To:    nextStation.Name,
					Price: 0, // Will be updated in second pass
				})
			}
		}

		log.Printf("Total known price: %f, Unknown segments: %d, Total price: %f", totalKnownPrice, unknownSegments, route.Price)

		// Calculate price for unknown segments
		remainingPrice := route.Price - totalKnownPrice
		if unknownSegments > 0 {
			pricePerUnknownSegment := remainingPrice / float64(unknownSegments)
			log.Printf("Remaining price: %f, Price per unknown segment: %f", remainingPrice, pricePerUnknownSegment)
			// Second pass: Update unknown segment prices
			for i := range legs {
				if legs[i].Price == 0 {
					legs[i].Price = pricePerUnknownSegment
				}
			}
		}

		// Validate total price matches
		var calculatedTotal float64
		for _, leg := range legs {
			calculatedTotal += leg.Price
		}
		if calculatedTotal != route.Price {
			log.Printf("Warning: Calculated total (%f) does not match route price (%f)", calculatedTotal, route.Price)
		}

		return &Journey{
			Stations:   stations,
			TotalPrice: route.Price,
			Legs:       legs,
		}, nil
	}

	return nil, fmt.Errorf("invalid route configuration")
}
