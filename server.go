package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

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

    log.Printf("Received a payment request: Phone Number=%v, Amount=%v", in.PhoneNumber, in.Amount)

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
	
	// Start the database and use the collection
	client, err := db.StartDatabase(ctx)  // Ensure this function is exported in db package
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")
	
	// Adding a new user
    newUser := map[string]interface{}{
        "_id": "user123",
        "name": "John Doe",
        "balance": 100.0,
    }
    err = db.AddUser(ctx, client, newUser)
    if err != nil {
        log.Fatalf("Failed to add user: %v", err)
    }
    log.Println("User added successfully")

    // Updating a user's balance
    err = db.UpdateUserBalance(ctx, client, "user123", 150.0)
    if err != nil {
        log.Fatalf("Failed to update user balance: %v", err)
    }
    log.Println("User balance updated successfully")


	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{})
	log.Println("Server listening at", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
