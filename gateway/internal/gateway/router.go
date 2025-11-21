package gateway

import (
	"net/http"
	"time"

	"rate-limiter-api-gateway/internal/gateway/middleware"
	pb "rate-limiter-api-gateway/proto"

	"github.com/gin-gonic/gin"
)

// NewRouter constructs the HTTP router and wires middlewares.
// rlClient is a gRPC client to the RateLimiter service.
func NewRouter(rlClient pb.RateLimiterClient) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.LoggingMiddleware())			// log all requests
	r.Use(middleware.JWTAuthMiddleware())			// authn via JWT
	r.Use(middleware.RateLimitMiddleware(rlClient))	// rate limiting

	// health check endpoint
	r.GET("/api/ping", ping)

	return r
}

func ping(c *gin.Context) {
	type resp struct {
		Message string `json:"message"`
		Time    string `json:"time"`
	}
	out := resp{Message: "pong", Time: time.Now().Format(time.RFC3339)}
	c.JSON(http.StatusOK, out)
}
