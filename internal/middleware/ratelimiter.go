package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	visits = make(map[string][]time.Time)
	mu     sync.Mutex
)

// RateLimiter sets a maximum threshold of requests permitted over a static time window per IP.
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()

		var valid []time.Time

		// Purge old requests outside the target window
		for _, t := range visits[ip] {
			if now.Sub(t) < window {
				valid = append(valid, t)
			}
		}

		// Check capacity
		if len(valid) >= maxRequests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		valid = append(valid, now)
		visits[ip] = valid
		mu.Unlock()

		c.Next()
	}
}
