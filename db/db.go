package db

import (
	"context"
	"fmt"
	"os"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
    ID       string `bson:"_id,omitempty"`
    Username string `bson:"username"`
    Email    string `bson:"email"`
    Password string `bson:"password"` // This should be a hashed password
	Balance float64 `bson:"balance"`
}

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

func GetUserByID(ctx context.Context, client *mongo.Client, id string) (*User, error) {
    // Getting a handle for the users collection in your database
    usersCollection := client.Database("yourDatabaseName").Collection("users")

    // Define a variable to hold the user info
    var user User

    // Finding a single document
    err := usersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("no user found with id %s: %w", id, err)
        }
        return nil, fmt.Errorf("error fetching user by id %s: %w", id, err)
    }

    return &user, nil
}

func AddUser(ctx context.Context, client *mongo.Client, user User) error {
    usersCollection := client.Database("yourDatabaseName").Collection("users")
    _, err := usersCollection.InsertOne(ctx, user)
    if err != nil {
        return fmt.Errorf("failed to add user: %w", err)
    }
    return nil
}

func UpdateUserBalanceWithTransaction(ctx context.Context, client *mongo.Client, userID string, amountToDeduct float64) error {
    session, err := client.StartSession()
    if err != nil {
        return fmt.Errorf("failed to start session: %w", err)
    }
    defer session.EndSession(ctx)

    transactionErr := mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) (interface{}, error) {
        usersCollection := client.Database("yourDatabaseName").Collection("users")

        var user struct {
            Balance float64 `bson:"balance"`
        }
        if err := usersCollection.FindOne(sessionContext, bson.M{"_id": userID}).Decode(&user); err != nil {
            return nil, fmt.Errorf("failed to fetch user balance: %w", err)
        }

        if user.Balance < amountToDeduct {
            return nil, fmt.Errorf("insufficient funds")
        }

        newBalance := user.Balance - amountToDeduct
        update := bson.M{"$set": bson.M{"balance": newBalance}}
        _, err = usersCollection.UpdateOne(sessionContext, bson.M{"_id": userID}, update)
        if err != nil {
            return nil, fmt.Errorf("failed to update user balance: %w", err)
        }

        return nil, nil
    })

    if transactionErr != nil {
        return fmt.Errorf("transaction failed: %w", transactionErr)
    }

    return nil
}