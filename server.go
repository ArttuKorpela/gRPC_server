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
)

type server struct {
	pb.UnimplementedPaymentServiceServer
}

func (s *server) ProcessPayment(ctx context.Context, in *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	log.Printf("Received a payment request: %v", in)

	// Construct the request to the Node.js server
	response, err := makeHTTPRequest(ctx, in.Phone_number, in.Amount)
	if err != nil {
		log.Printf("Error calling Node.js server: %v", err)
		return nil, err
	}

	// Log and use the response from the Node.js server
	log.Printf("Node.js server response: %s", response)
	accepted := response == "Payment confirmed by user"

	// Return the response based on the Node.js confirmation
	return &pb.PaymentResponse{Accepted: accepted}, nil
}

func makeHTTPRequest(ctx context.Context, phone_number int32, amount float64) (string, error) {
	// Marshal the data into a JSON payload
	payload, err := json.Marshal(map[string]interface{}{
		"phonenumber": phone_number,
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

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{})
	log.Println("Server listening at", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}