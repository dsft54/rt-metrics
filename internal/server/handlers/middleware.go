package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
func Compression() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
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

func Decryption(keyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyData, err := ioutil.ReadFile(keyPath)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		block, _ := pem.Decode(keyData)
		private, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		log.Println(string(body))
		msgLen := len(body)
		step := 384
		var decryptedBytes []byte
		for start := 0; start < msgLen; start += step {
			finish := start + step
			if finish > msgLen {
				finish = msgLen
			}
			decryptedBlock, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, private.(*rsa.PrivateKey), body[start:finish], nil)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			decryptedBytes = append(decryptedBytes, decryptedBlock...)
		}
		log.Println(string(decryptedBytes))
		c.Request.ContentLength = int64(len(decryptedBytes))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedBytes))
		c.Next()
	}
}
