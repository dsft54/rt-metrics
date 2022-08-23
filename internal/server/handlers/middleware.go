package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	
	"github.com/dsft54/rt-metrics/internal/cryptokey"
)

type gzipBodyWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

// Write подготавливает writer для его передачи после обработчика.
func (gz gzipBodyWriter) Write(b []byte) (int, error) {
	return gz.writer.Write(b)
}

// Compression middleware - сжимает тело запроса/ответа и передает дальше по цепочке обработчиков.
func Compression(gzipSpeed int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzipSpeed)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer gz.Close()

		c.Writer = gzipBodyWriter{c.Writer, gz}
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Next()
	}
}

// Decompression middleware - распаковывает тело запроса/ответа и передает дальше по цепочке обработчиков.
func Decompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") ||
			!strings.Contains(c.Request.Header.Get("Content-Encoding"), "deflate") ||
			!strings.Contains(c.Request.Header.Get("Content-Encoding"), "br") {
			c.Next()
			return
		}

		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer gz.Close()

		body, err := io.ReadAll(gz)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Request.ContentLength = int64(len(body))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Next()
	}
}

func Decryption(private *rsa.PrivateKey, chunkSize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		log.Println(string(body))
		decryptedBody, err := cryptokey.DecryptMessage(body, private, chunkSize)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		log.Println(string(decryptedBody))
		c.Request.ContentLength = int64(len(decryptedBody))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
		c.Next()
	}
}
