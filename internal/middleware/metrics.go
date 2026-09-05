package middleware

import (
	"strconv"
	"time"

	"github.com/Leongt1/url-shortener/internal/metrics"
	"github.com/gin-gonic/gin"
)

func RequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // runs before the handler

		c.Next()

		elapsedSeconds := time.Since(start).Seconds() // runs after the handler

		path := c.FullPath()
		if len(path) == 0 {
			path = "unmatched"
		}

		metrics.RequestDuration.
			WithLabelValues(c.Request.Method, path, strconv.Itoa(c.Writer.Status())).
			Observe(elapsedSeconds)
	}
}
