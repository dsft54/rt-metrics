package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	type args struct {
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Normal conditions",
			args: args{
				metricType:  "gauge",
				metricName:  "Alloc",
				metricValue: "3.14159265",
			},
			want:    200,
			wantErr: false,
		},
		{
			name: "Failed to parse gauge float",
			args: args{
				metricType:  "gauge",
				metricName:  "Alloc",
				metricValue: "3.AAAA1415",
			},
			want:    400,
			wantErr: true,
		},
		{
			name: "Failed to parse counter int",
			args: args{
				metricType:  "counter",
				metricName:  "PollCounter",
				metricValue: "3AAA",
			},
			want:    400,
			wantErr: true,
		},
		{
			name: "Unknown metric type",
			args: args{
				metricType:  "unknown",
				metricName:  "Alloc",
				metricValue: "3.14159265",
			},
			want:    501,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateMetrics(tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UpdateMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatesHandler(t *testing.T) {

	tests := []struct {
		name   string
		method string
		url    string
		code   int
	}{
		{
			name:   "Normal conditions",
			method: "POST",
			url:    "/update/gauge/Alloc/3.14159265",
			code:   200,
		},
		{
			name:   "Failed to parse",
			method: "POST",
			url:    "/update/gauge/Alloc/3.AAA14159265",
			code:   400,
		},
		{
			name:   "Wrong type",
			method: "POST",
			url:    "/update/wrong/Alloc/3.14159265",
			code:   501,
		},
		{
			name:   "Wrong method",
			method: "GET",
			url:    "/update/gauge/Alloc/3.14159265",
			code:   404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/update/:type/:name/:value", UpdatesHandler)
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestAddressedRequest(t *testing.T) {
	gaugeMetrics = map[string]float64{"Alloc": 3.14159265}
	tests := []struct {
		name   string
		method string
		url    string
		code   int
	}{
		{
			name:   "Normal conditions",
			method: "GET",
			url:    "/value/gauge/Alloc",
			code:   200,
		},
		{
			name:   "Not existing field",
			method: "GET",
			url:    "/value/gauge/Al",
			code:   404,
		},
		{
			name:   "Wrong method",
			method: "POST",
			url:    "/update/gauge/All/3.14159265",
			code:   404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/value/:type/:name", AddressedRequest)
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}
