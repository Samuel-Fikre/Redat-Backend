package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func ConnectDB(mongoURI string) error {
	log.Println("Attempting to connect to MongoDB...")

	// Set client options with longer timeout and retry logic
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetSocketTimeout(10 * time.Second)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("✅ Successfully connected to MongoDB!")
	return nil
}

func GetCollection(dbName string, collectionName string) *mongo.Collection {
	collection := mongoClient.Database(dbName).Collection(collectionName)
	return collection
}

func DisconnectDB() {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("❌ Error disconnecting from MongoDB: %v", err)
			return
		}
		log.Println("✅ Successfully disconnected from MongoDB")
	}
}
