package middleware

import (
	"evo-ai-core-service/internal/httpclient/response"
	"evo-ai-core-service/internal/utils/contextutils"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for different clients
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// getLimiter returns the rate limiter for a specific client
func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientID]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[clientID] = limiter
	}

	return limiter
}

// cleanupOldLimiters removes unused limiters to prevent memory leaks
func (rl *RateLimiter) cleanupOldLimiters() {
	ticker := time.NewTicker(time.Minute * 5) // cleanup every 5 minutes
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			// Remove limiters that haven't been used recently
			for clientID, limiter := range rl.limiters {
				// If the limiter has tokens available and hasn't been used recently, remove it
				if limiter.Tokens() == float64(rl.burst) {
					delete(rl.limiters, clientID)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	rateLimiter := NewRateLimiter(requestsPerSecond, burst)
	rateLimiter.cleanupOldLimiters()

	return func(c *gin.Context) {
		clientID := fmt.Sprintf("ip_%s", c.ClientIP()) // Default fallback

		if userID, err := contextutils.GetUserID(c.Request.Context()); err == nil {
			clientID = fmt.Sprintf("user_%s", userID)
		}

		limiter := rateLimiter.getLimiter(clientID)

		if !limiter.Allow() {
			response.ErrorResponse(c, "ERR_RATE_LIMIT_EXCEEDED", "Rate limit exceeded", nil, http.StatusTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimitMiddleware creates a global rate limiting middleware
func GlobalRateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.ErrorResponse(c, "ERR_RATE_LIMIT_EXCEEDED", "Global rate limit exceeded", nil, http.StatusTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
	}
}
