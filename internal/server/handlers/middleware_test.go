package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dsft54/rt-metrics/internal/cryptokey"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert"
)

func TestCompression(t *testing.T) {
	req1, _ := http.NewRequest("POST", "/", nil)
	req2, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte("message")))
	req2.Header.Set("Accept-Encoding", "gzip")
	tests := []struct {
		name  string
		req   *http.Request
		speed int
		want  int
	}{
		{
			name:  "normal",
			req:   req1,
			speed: gzip.BestSpeed,
			want:  200,
		},
		{
			name:  "gzip err",
			req:   req2,
			speed: 10,
			want:  500,
		},
		{
			name:  "gzip clean",
			req:   req2,
			speed: gzip.BestSpeed,
			want:  200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.Use(Compression(tt.speed))
			r.POST("/", func(c *gin.Context) {
				c.Status(200)
			})
			r.ServeHTTP(w, tt.req)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestDecompression(t *testing.T) {
	req1, _ := http.NewRequest("POST", "/", nil)
	req2, _ := http.NewRequest("POST", "/", errReader(0))
	req2.Header.Add("Content-Encoding", "deflate, gzip, br")
	req3, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte{31, 139, 8, 0, 0, 0, 0, 0, 4, 255, 1, 0, 0, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0}))
	req3.Header.Add("Content-Encoding", "deflate, gzip, br")
	tests := []struct {
		name string
		req  *http.Request
		want int
	}{
		{
			name: "normal",
			req:  req1,
			want: 200,
		},
		{
			name: "reader err",
			req:  req2,
			want: 500,
		},
		{
			name: "gz ok",
			req:  req3,
			want: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.Use(Decompression())
			r.POST("/", func(c *gin.Context) {
				c.Status(200)
			})
			r.ServeHTTP(w, tt.req)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestDecryption(t *testing.T) {
	priv, err := cryptokey.ParsePrivateKey("keytest")
	if err != nil {
		t.Error(err)
	}
	req1, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte{}))
	req2, _ := http.NewRequest("POST", "/", errReader(0))
	req3, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte{31, 139, 8, 0, 0, 0, 0, 0, 4, 255, 1, 0, 0, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0}))
	tests := []struct {
		name      string
		private   *rsa.PrivateKey
		req       *http.Request
		chunkSize int
		want      int
	}{
		{
			name:      "normal",
			private:   priv,
			chunkSize: 384,
			req:       req1,
			want:      200,
		},
		{
			name:      "normal",
			private:   priv,
			chunkSize: 384,
			req:       req2,
			want:      500,
		},
		{
			name:      "normal",
			private:   priv,
			chunkSize: 384,
			req:       req3,
			want:      500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.Use(Decryption(tt.private, tt.chunkSize))
			r.POST("/", func(c *gin.Context) {
				c.Status(200)
			})
			r.ServeHTTP(w, tt.req)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}
