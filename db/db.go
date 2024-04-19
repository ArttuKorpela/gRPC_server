package db

import (
	"context"
	"fmt"

	os "github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StartDatabase initializes a MongoDB client and returns it.
// It returns a *mongo.Client and any error encountered.
func StartDatabase(ctx context.Context) (*mongo.Client, error) {
	// Load .env file
    err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
    }
	// Get the MongoDB URI from environment variables
	mongoURI := os.Getenv("MONGO_URI")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		client.Disconnect(ctx)  // Disconnect only if there is an error after connection
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client, nil
}