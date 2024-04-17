package db

import (
	"context"


	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// startDatabase initializes a MongoDB client and returns a reference to the users collection.
// It returns a mongo.Collection pointer and any error encountered.
func StartDatabase(ctx context.Context) (*mongo.Collection, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)  // Attempt to disconnect the client to cleanup resources
		return nil, err
	}

	usersCollection := client.Database("testing").Collection("users")
	return usersCollection, nil
}