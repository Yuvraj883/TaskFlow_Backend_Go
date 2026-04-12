package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLogger logs HTTP requests using slog
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()

		latency := time.Since(start)
		
		slog.Info("HTTP Request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.String("latency", latency.String()),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
