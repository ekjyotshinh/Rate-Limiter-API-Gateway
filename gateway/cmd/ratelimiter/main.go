package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"rate-limiter-api-gateway/internal/ratelimiter"
	pb "rate-limiter-api-gateway/proto"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedRateLimiterServer
	rdb    *redis.Client
	luaSHA string
}

func (s *server) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	key := req.Key
	now := req.NowMs
	// Note: convert arguments to interfaces for EvalSha
	args := []interface{}{now, req.Capacity, req.RefillRate, req.Cost, req.TtlSeconds}

	res, err := s.rdb.EvalSha(ctx, s.luaSHA, []string{key}, args...).Result()
	if err != nil {
		// If script missing, reload and retry once (simple recovery).
		if err == redis.Nil || err.Error() == "NOSCRIPT No matching script. Please use EVAL." {
			sha, loadErr := ratelimiter.LoadLuaScript(ctx, s.rdb, "internal/ratelimiter/lua/token_bucket.lua")
			if loadErr == nil {
				s.luaSHA = sha
				res, err = s.rdb.EvalSha(ctx, s.luaSHA, []string{key}, args...).Result()
			}
		}
		if err != nil {
			return nil, err
		}
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 3 {
		return nil, fmt.Errorf("unexpected evalsha result: %v", res)
	}

	allowedInt, _ := strconv.ParseInt(fmt.Sprintf("%v", arr[0]), 10, 64)
	allowed := allowedInt == 1
	tokensLeft, _ := strconv.ParseFloat(fmt.Sprintf("%v", arr[1]), 64)
	capacity, _ := strconv.ParseFloat(fmt.Sprintf("%v", arr[2]), 64)

	return &pb.CheckResponse{
		Allowed:    allowed,
		TokensLeft: tokensLeft,
		Capacity:   capacity,
	}, nil
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // default for development
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	// load lua
	sha, err := ratelimiter.LoadLuaScript(ctx, rdb, "internal/ratelimiter/lua/token_bucket.lua")
	if err != nil {
		log.Fatalf("failed to load lua script: %v", err)
	}

	// create gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRateLimiterServer(s, &server{rdb: rdb, luaSHA: sha})
	log.Println("ratelimiter gRPC listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC serve error: %v", err)
	}
}
