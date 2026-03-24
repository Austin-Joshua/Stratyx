package http

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	LatencyMS int64  `json:"latencyMs"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("requestId", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get("requestId")
		entry := logEntry{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     "info",
			Message:   "http_request",
			RequestID: toString(requestID),
			Method:    c.Request.Method,
			Path:      c.FullPath(),
			Status:    c.Writer.Status(),
			LatencyMS: time.Since(start).Milliseconds(),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		if entry.Path == "" {
			entry.Path = c.Request.URL.Path
		}
		if entry.Status >= 500 {
			entry.Level = "error"
		}
		payload, err := json.Marshal(entry)
		if err != nil {
			log.Printf(`{"level":"error","message":"log_marshal_failed","error":"%s"}`, err.Error())
			return
		}
		log.Println(string(payload))
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
