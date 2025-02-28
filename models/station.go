package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Location struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // [longitude, latitude]
}

type Station struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name            string             `json:"name" bson:"name"`
	Image           string             `json:"image" bson:"image,omitempty"`
	Location        Location           `json:"location" bson:"location"`
	ConnectedRoutes []string           `json:"connected_routes" bson:"connected_routes"`
}

func (s *Station) CreateGeospatialIndex(collection *mongo.Collection) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "location", Value: "2dsphere"},
		},
	}
	_, err := collection.Indexes().CreateOne(context.Background(), index)
	return err
}
