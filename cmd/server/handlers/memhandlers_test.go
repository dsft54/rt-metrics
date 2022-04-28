package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dsft54/rt-metrics/cmd/server/storage"
)

func TestAddressedRequest(t *testing.T) {
	testStore := storage.MemoryStorage{}
	testStore.GaugeMetrics = map[string]float64{"Alloc": 3.14159265}
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
			url:    "/value/gauge/All/3.14159265",
			code:   404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/value/:type/:name", AddressedRequest(&testStore))
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestHandleRequestJSON(t *testing.T) {
	testStore := storage.MemoryStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
	tests := []struct {
		name         string
		request      storage.Metrics
		response     storage.Metrics
		code         int
		wantNilDelta bool
		wantNilValue bool
	}{
		{
			name: "Normal Request",
			request: storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			response:     storage.Metrics{},
			code:         200,
			wantNilDelta: true,
			wantNilValue: false,
		},
		{
			name: "Wrong Type",
			request: storage.Metrics{
				ID:    "Alloc",
				MType: "wrong",
			},
			response:     storage.Metrics{},
			code:         200,
			wantNilDelta: true,
			wantNilValue: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/value", HandleRequestJSON(&testStore, ""))

			dataReq, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf(err.Error())
			}

			req, _ := http.NewRequest("POST", "/value", bytes.NewBuffer(dataReq))
			r.ServeHTTP(w, req)
			err = json.Unmarshal(w.Body.Bytes(), &tt.response)
			if err != nil {
				t.Errorf(err.Error())
			}

			DeltaISNil, ValueISNil := false, false
			if tt.response.Delta == nil {
				DeltaISNil = true
			}
			if tt.response.Value == nil {
				ValueISNil = true
			}

			assert.Equal(t, DeltaISNil, tt.wantNilDelta)
			assert.Equal(t, ValueISNil, tt.wantNilValue)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestHandleUpdateJSON(t *testing.T) {
	testStore := storage.MemoryStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
	testFileStore := storage.FileStorage{
		Synchronize: false,
		StoreData:   false,
	}
	tests := []struct {
		name    string
		valueG  float64
		valueC  int64
		request storage.Metrics
		code    int
		wantG   float64
		wantC   int64
	}{
		{
			name:   "Normal Gauge Update Request",
			valueG: 31.123,
			request: storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			code:  200,
			wantG: 31.123,
			wantC: 0,
		},
		{
			name:   "Normal Counter Update Request",
			valueC: 123,
			request: storage.Metrics{
				ID:    "PollCounter",
				MType: "counter",
			},
			code:  200,
			wantG: 0,
			wantC: 123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/update", HandleUpdateJSON(&testStore, &testFileStore,  ""))

			if tt.request.MType == "gauge" {
				tt.request.Value = &tt.valueG
			}
			if tt.request.MType == "counter" {
				tt.request.Delta = &tt.valueC
			}
			dataReq, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf(err.Error())
			}

			req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(dataReq))
			r.ServeHTTP(w, req)

			assert.Equal(t, testStore.GaugeMetrics[tt.request.ID], tt.wantG)
			assert.Equal(t, testStore.CounterMetrics[tt.request.ID], tt.wantC)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}
