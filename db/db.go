package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

// startDatabase initializes a MongoDB client and returns a reference to the users collection.
// It returns a mongo.Collection pointer and any error encountered.
func StartDatabase(ctx context.Context) (*mongo.Collection, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://arttukorpela01:testi123!@distributedsystemstesti.a3bqykr.mongodb.net/?retryWrites=true&w=majority&appName=DistributedSystemsTesting").SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
		  panic(err)
		}
	  }()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
	panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client.Database("users").Collection("documents"), err
}