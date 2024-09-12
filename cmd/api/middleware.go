package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	maxRequestBodySize = 1 << 20
)

// Middleware to limit request body size
func setMaxSizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxRequestBodySize))
		c.Next()
	}
}
