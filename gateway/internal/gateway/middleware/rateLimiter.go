package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	pb "rate-limiter-api-gateway/proto"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// RateLimitMiddleware returns a middleware that calls the RateLimiter gRPC service.
func RateLimitMiddleware(rlClient pb.RateLimiterClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// build key: prefer user-sub (if present), else use IP
		var key string
		if tokenVal, exists := c.Get("jwt"); exists {
			if token, ok := tokenVal.(*jwt.Token); ok {
				claims := token.Claims.(jwt.MapClaims)
				if sub, found := claims["sub"].(string); found && sub != "" {
					key = fmt.Sprintf("user:%s:endpoint:%s", sub, c.Request.URL.Path)
				}
			}
		}
		// fallback to IP-based key -- should not happen if auth is used
		if key == "" {
			key = fmt.Sprintf("ip:%s:endpoint:%s", c.ClientIP(), c.Request.URL.Path)
		}

		now := time.Now().UnixNano() / int64(time.Millisecond)
		req := &pb.CheckRequest{
			Key:        key,
			NowMs:      now,
			Capacity:   10,    // capacity per endpoint per user -tune per endpoint use config
			RefillRate: 1.0,   // 1 token/second - tune per endpoint use config
			Cost:       1,     // 1 token per request - tune per endpoint use config
			TtlSeconds: 3600,  // keep counters an hour - tune per endpoint use config
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		resp, err := rlClient.Check(ctx, req)
		if err != nil {
			// On error, choose fail-open for availability. -- most application fail close to prevent abuse.
			log.Printf("ratelimiter gRPC error: %v", err)
			c.Next()
			return
		}
		if !resp.Allowed {
			c.Header("Retry-After", "1") // configurable use config
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		// Attach headers for observability
		c.Header("X-RateLimit-Limit", strconv.FormatFloat(resp.Capacity, 'f', 0, 64))
		c.Header("X-RateLimit-Remaining", strconv.FormatFloat(resp.TokensLeft, 'f', 0, 64))

		c.Next()
	}
}
