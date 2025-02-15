package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Route struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	From                 string             `json:"from" bson:"from"`
	To                   string             `json:"to" bson:"to"`
	Price                float64            `json:"price" bson:"price"`
	IsDirectRoute        bool               `json:"isDirectRoute" bson:"isDirectRoute"`
	IntermediateStations []string           `json:"intermediateStations,omitempty" bson:"intermediateStations,omitempty"`
}

type JourneyResponse struct {
	Route      []string   `json:"route"`
	TotalPrice float64    `json:"totalPrice"`
	Legs       []RouteLeg `json:"legs"`
}

type RouteLeg struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Price float64 `json:"price"`
}
