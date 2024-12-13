package middleware

import (
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func DecompressRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Method == http.MethodGet {
			c.Next()
			return
		}
		switch c.GetHeader("Content-Encoding") {
		case "gzip":
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			// Replace the request body with the decompressed data
			c.Request.Body = io.NopCloser(gzipReader)
			c.Request.Header.Del("Content-Encoding")
		case "br":
			reader := brotli.NewReader(c.Request.Body)
			c.Request.Body = io.NopCloser(reader)
			c.Request.Header.Del("Content-Encoding")
		}

		// Continue processing the request
		c.Next()
	}
}
