package main


import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/ArttuKorpela/gRPC_server/server/payment"
)

type server struct {
	pb.UnimplementedPaymentServiceServer
}

func (s *server) PayOrder(ctx context.Context, in *pb.Payment) (*pb.Response, error) {
	log.Printf("Got a request: %v", in.Ammount)
	return &pb.Response{Status: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err!= nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{})
	if err := s.Serve(lis); err!= nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
