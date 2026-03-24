package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type limitEntry struct {
	count       int
	windowStart time.Time
}

var (
	rateMu      sync.Mutex
	rateBuckets = map[string]*limitEntry{}
)

func AuthRateLimit() gin.HandlerFunc {
	const (
		maxRequests = 20
		window      = time.Minute
	)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rateMu.Lock()
		entry, ok := rateBuckets[ip]
		if !ok || now.Sub(entry.windowStart) > window {
			entry = &limitEntry{count: 0, windowStart: now}
			rateBuckets[ip] = entry
		}
		entry.count++
		allowed := entry.count <= maxRequests
		rateMu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
