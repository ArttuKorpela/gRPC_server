package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"


	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	pb "github.com/ArttuKorpela/gRPC_server/server/payment"
	db "github.com/ArttuKorpela/gRPC_server/server/db"
)

type server struct {
	pb.UnimplementedPaymentServiceServer
}



func (s *server) ProcessPayment(ctx context.Context, in *pb.Payment) (*pb.Response, error) {
	type JSONResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

    log.Printf("Received a payment request: Phone Number=%s, Amount=%v", in.PhoneNumber, in.Amount)
	if in.PhoneNumber == "" {
        log.Println("PhoneNumber is empty")
    } else {
        log.Printf("Logging PhoneNumber again: %s", in.PhoneNumber)
    }
    // Construct the request to the Node.js server
    response, err := makeHTTPRequest(ctx, in.PhoneNumber, in.Amount)
    if err != nil {
        log.Printf("Error calling Node.js server: %v", err)
        return nil, err
    }

    // Log and use the response from the Node.js server
    log.Printf("Node.js server response: %s", response)
	var jsonResponse JSONResponse
	json.Unmarshal([]byte(response), &jsonResponse)
	log.Printf("%s", jsonResponse.Message)
    accepted := jsonResponse.Message == "Payment confirmation received"
	log.Printf("%b",accepted)
    // Return the response based on the Node.js confirmation
    return &pb.Response{Accepted: accepted}, nil
}

func makeHTTPRequest(ctx context.Context, phoneNumber string, amount float64) (string, error) {
    // Marshal the data into a JSON payload
	log.Printf("Phone: %s", phoneNumber)
    payload, err := json.Marshal(map[string]interface{}{
        "phoneNumber": phoneNumber,
        "amount": amount,
    })
    if err != nil {
        return "", err
    }

    // Create a new request with the payload
    req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8000/confirm-payment", bytes.NewBuffer(payload))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Read and return the response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}


func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Correct function name and removed unused 'cancel'
	ctx := context.Background()

	err = godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file")
    }
	// Get the MongoDB URI from environment variables
	
	mongoURI := os.Getenv("MONGO_URI")
	
	// Start the database and use the collection
	client, err := db.StartDatabase(ctx, mongoURI)  // Ensure this function is exported in db package
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	r := gin.Default()

	r.POST("/users", func(c *gin.Context) {
		

		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
	
		// Log the body content
		bodyString := string(bodyBytes)
		log.Println("Request Body:", bodyString)
	
		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	
		// Proceed to bind JSON as usual
		var newUser db.User
		if err := c.BindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}
		log.Println("User added", newUser)
		err = db.AddUser(ctx, client, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"userId": newUser.ID, 
		})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		user, err := db.GetUserByID(ctx, client, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	r.POST("/users/:id/balance", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			NewBalance float64 `json:"newBalance"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := db.UpdateUserBalance(ctx, client, id, req.NewBalance)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})
	
	
	//gRPC thread
	s := grpc.NewServer()
    pb.RegisterPaymentServiceServer(s, &server{})
    go func() {
        log.Println("gRPC server listening at", lis.Addr())
        if err := s.Serve(lis); err != nil {
            log.Fatalf("failed to serve gRPC: %v", err)
        }
    }()
	//HTTP thread
	go func() {
        log.Println("HTTP server running on localhost:5000")
        if err := r.Run("localhost:5000"); err != nil {
            log.Fatalf("failed to run HTTP server: %v", err)
        }
    }()

	// Block the main thread, or wait for a signal to gracefully shutdown
    select {}
}


/*
	newUser:= db.User{
		ID: "test",
		Username: "john_doe",
        Email:    "john.doe@example.com",
        Password: "hashed_password_here",
		Balance: 70.0,
	}
	
    err = db.AddUser(ctx, client, newUser)
    if err != nil {
        log.Fatalf("Failed to add user: %v", err)
    }
    log.Println("User added successfully")
	*/
    // Updating a user's balance
	/*
    err = db.UpdateUserBalance(ctx, client, "test", 125.0)
    if err != nil {
        log.Fatalf("Failed to update user balance: %v", err)
    }
    log.Println("User balance updated successfully")

 	user, err:= db.GetUserByID(ctx,client,"test")
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}
	log.Println("Username:",user.Username)
	*/