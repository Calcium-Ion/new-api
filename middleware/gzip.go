package middleware

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func GzipDecodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			// Replace the request body with the decompressed data
			c.Request.Body = io.NopCloser(gzipReader)
		}

		// Continue processing the request
		c.Next()
	}
}
