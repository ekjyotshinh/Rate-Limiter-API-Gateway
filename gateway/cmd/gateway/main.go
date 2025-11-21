package main

import (
	"log"
	"os"
	"time"

	"rate-limiter-api-gateway/internal/gateway"
	pb "rate-limiter-api-gateway/proto"

	"google.golang.org/grpc"
)

func main() {
	rlAddr := os.Getenv("RATELIMITER_ADDR")
	if rlAddr == "" {
		rlAddr = "localhost:50051"
	}

	// Dial the ratelimiter gRPC service (insecure for local demo)
	conn, err := grpc.Dial(rlAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		log.Fatalf("failed to dial ratelimiter: %v", err)
	}
	defer conn.Close()
	rlClient := pb.NewRateLimiterClient(conn)

	// Build router
	r := gateway.NewRouter(rlClient)

	log.Println("gateway listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
