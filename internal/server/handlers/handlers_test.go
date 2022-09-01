package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert"

	"github.com/dsft54/rt-metrics/internal/server/storage"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

var (
	errNotFound  = fmt.Errorf("metric not found")
	errWrongType = fmt.Errorf("wrong metric type")
)

type mockStorage struct {
}

func (ms *mockStorage) ParamsUpdate(metricType, metricName, metricValue string) (int, error) {
	if metricType == "gauge" {
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return 400, err
		}
		return 200, nil
	}
	if metricType == "counter" {
		_, err := strconv.Atoi(metricValue)
		if err != nil {
			return 400, err
		}
		return 200, nil
	}
	return 501, errWrongType
}

func (ms *mockStorage) InsertMetric(m *storage.Metrics) error {
	if m.ID == "ERROR" {
		return errNotFound
	}
	return nil
}

func (ms *mockStorage) InsertBatchMetric(m []storage.Metrics) error {
	if len(m) != 0 && m[0].ID == "ERROR" {
		return errNotFound
	}
	return nil
}

func (ms *mockStorage) ReadMetric(mr *storage.Metrics) (*storage.Metrics, error) {
	var (
		value float64
		delta int64
	)
	switch mr.MType {
	case "gauge":
		if mr.ID != "Alloc" {
			return nil, errNotFound
		}
		mr.Value = &value
		return mr, nil
	case "counter":
		if mr.ID != "Pollcount" {
			return nil, errNotFound
		}
		mr.Delta = &delta
		return mr, nil
	}
	return nil, errNotFound
}

func (ms *mockStorage) ReadAllMetrics() ([]storage.Metrics, error) {
	var (
		delta int64
		value float64
	)
	mmSlice := []storage.Metrics{
		{ID: "Alloc",
			MType: "gauge",
			Value: &value},
		{ID: "Counter",
			MType: "counter",
			Delta: &delta},
	}
	return mmSlice, nil
}

func (ms *mockStorage) SaveToFile(f *os.File) error {
	_, err := f.Write([]byte("mock_test"))
	if err != nil {
		return err
	}
	return nil
}

func (ms *mockStorage) UploadFromFile(string) error {
	return nil
}

func (ms *mockStorage) Ping() error {
	return nil
}

func TestParametersUpdate(t *testing.T) {
	tests := []struct {
		name string
		st   *mockStorage
		fs   *storage.FileStorage
		url  string
		code int
	}{
		{
			name: "Basic correct params test",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "",
				Synchronize: false,
			},
			url:  "/update/gauge/alloc/3.1415",
			code: 200,
		},
		{
			name: "Unexpected float test",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "",
				Synchronize: false,
			},
			url:  "/update/counter/alloc/3.1415",
			code: 400,
		},
		{
			name: "Wrong type test",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "",
				Synchronize: false,
			},
			url:  "/update/wrongtype/alloc/3.1415",
			code: 501,
		},
		{
			name: "Sync: basic correct save to file",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			url:  "/update/gauge/alloc/3.1415",
			code: 200,
		},
		{
			name: "Sync: empty file name",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "",
				Synchronize: true,
			},
			url:  "/update/gauge/alloc/3.1415",
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/update/:type/:name/:value", ParametersUpdate(tt.st, tt.fs))
			req, _ := http.NewRequest("POST", tt.url, nil)
			r.ServeHTTP(w, req)
			if tt.fs.Synchronize && tt.fs.FilePath != "" {
				data, err := ioutil.ReadFile(tt.fs.FilePath)
				if err != nil {
					t.Error(err, tt)
				}
				err = os.Remove(tt.fs.FilePath)
				if err != nil {
					t.Error(err, tt)
				}
				assert.Equal(t, string(data), "mock_test")
			}
			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestAddressedRequest(t *testing.T) {
	tests := []struct {
		name string
		st   *mockStorage
		url  string
		code int
	}{
		{
			name: "Existing metric gauge test",
			st:   &mockStorage{},
			url:  "/value/gauge/Alloc",
			code: 200,
		},
		{
			name: "Existing metric counter test",
			st:   &mockStorage{},
			url:  "/value/counter/Pollcount",
			code: 200,
		},
		{
			name: "Not existing metric test",
			st:   &mockStorage{},
			url:  "/value/wrong/type",
			code: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/value/:type/:name", AddressedRequest(tt.st))
			req, _ := http.NewRequest("GET", tt.url, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
			if w.Code == 200 {
				assert.Equal(t, w.Body.String(), "0")
			}
		})
	}
}

func TestRequestMetricJSON(t *testing.T) {
	var (
		v float64
		d int64
	)
	tests := []struct {
		st           *mockStorage
		request      *storage.Metrics
		wantResponse *storage.Metrics
		name         string
		key          string
		code         int
	}{
		{
			name: "Simple basic test",
			st:   &mockStorage{},
			key:  "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			wantResponse: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 200,
		},
		{
			name: "Err not found",
			st:   &mockStorage{},
			key:  "",
			request: &storage.Metrics{
				ID:    "Heap",
				MType: "counter",
			},
			wantResponse: &storage.Metrics{
				ID:    "Heap",
				MType: "counter",
				Value: &v,
			},
			code: 404,
		},
		{
			name: "Simple basic test with key",
			st:   &mockStorage{},
			key:  "testkey",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			wantResponse: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
				Hash:  "9764fd9e51f33bca83ea4218359b3e13257fac2c2ef33dfcb07e68625cfe02e5",
			},
			code: 200,
		},
		{
			name: "Simple basic test counter with key",
			st:   &mockStorage{},
			key:  "testkey",
			request: &storage.Metrics{
				ID:    "Pollcount",
				MType: "counter",
			},
			wantResponse: &storage.Metrics{
				ID:    "Pollcount",
				MType: "counter",
				Delta: &d,
				Hash:  "6c8781291e7d9d55b240dc460056f1661d5b4927119a5fa11c6a54eb1b3cd8e4",
			},
			code: 200,
		},
		{
			name: "reader err",
			st:   &mockStorage{},
			key:  "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			wantResponse: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 500,
		},
		{
			name: "unmarshall err",
			st:   &mockStorage{},
			key:  "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			wantResponse: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/value/", RequestMetricJSON(tt.st, tt.key))
			breq, err := json.Marshal(tt.request)
			if err != nil {
				t.Error(err)
			}
			switch tt.name {
			case "reader err":
				req, _ := http.NewRequest("POST", "/value/", errReader(0))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
			case "unmarshall err":
				req, _ := http.NewRequest("POST", "/value/", bytes.NewBuffer([]byte("message")))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
			default:
				req, _ := http.NewRequest("POST", "/value/", bytes.NewBuffer(breq))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
				if w.Code == 200 {
					jresp := &storage.Metrics{}
					err = json.Unmarshal(w.Body.Bytes(), jresp)
					if err != nil {
						t.Error(err)
					}
					assert.Equal(t, tt.wantResponse, jresp)
				}
			}
		})
	}
}

func TestUpdateMetricJSON(t *testing.T) {
	var (
		v float64
		d int64
	)
	tests := []struct {
		st      *mockStorage
		fs      *storage.FileStorage
		request *storage.Metrics
		name    string
		key     string
		code    int
	}{
		{
			name: "Simple basic test",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 200,
		},
		{
			name: "Simple basic test with key",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "testkey",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
				Hash:  "9764fd9e51f33bca83ea4218359b3e13257fac2c2ef33dfcb07e68625cfe02e5",
			},
			code: 200,
		},
		{
			name: "Simple basic test counter with key",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "testkey",
			request: &storage.Metrics{
				ID:    "Pollcount",
				MType: "counter",
				Delta: &d,
				Hash:  "6c8781291e7d9d55b240dc460056f1661d5b4927119a5fa11c6a54eb1b3cd8e4",
			},
			code: 200,
		},
		{
			name: "Wrong type",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "testkey",
			request: &storage.Metrics{
				ID:    "Pollcount",
				MType: "counteeeer",
				Delta: &d,
				Hash:  "6c8781291e7d9d55b240dc460056f1661d5b4927119a5fa11c6a54eb1b3cd8e4",
			},
			code: 500,
		},
		{
			name: "Insert err",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			request: &storage.Metrics{
				ID:    "ERROR",
				MType: "counter",
				Delta: &d,
			},
			code: 500,
		},
		{
			name: "Wrong hash",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "testkey",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
				Hash:  "3257fac2c2ef33dfcb07e68625cfe02e59764fd9e51f33bca83ea4218359b3e1",
			},
			code: 400,
		},
		{
			name: "Sync test",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			key: "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 200,
		},
		{
			name: "Sync test err",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "",
				Synchronize: true,
			},
			key: "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 500,
		},
		{
			name: "reader err",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			key: "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 500,
		},
		{
			name: "unmarshall err",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			key: "",
			request: &storage.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/update/", UpdateMetricJSON(tt.st, tt.fs, tt.key))
			breq, err := json.Marshal(tt.request)
			if err != nil {
				t.Error(err)
			}
			switch tt.name {
			case "reader err":
				req, _ := http.NewRequest("POST", "/update/", errReader(0))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
			case "unmarshall err":
				req, _ := http.NewRequest("POST", "/update/", bytes.NewBuffer([]byte{255, 1, 1, 2, 0, 0}))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
			default:
				req, _ := http.NewRequest("POST", "/update/", bytes.NewBuffer(breq))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
				if tt.fs.Synchronize && tt.fs.FilePath != "" {
					data, err := ioutil.ReadFile(tt.fs.FilePath)
					if err != nil {
						t.Error(err, tt)
					}
					err = os.Remove(tt.fs.FilePath)
					if err != nil {
						t.Error(err, tt)
					}
					assert.Equal(t, string(data), "mock_test")
				}
			}
		})
	}
}

func TestPingDatabase(t *testing.T) {
	tests := []struct {
		st   storage.IStorage
		name string
		code int
	}{
		{
			name: "Normal working ping",
			st:   &mockStorage{},
			code: 200,
		},
		{
			name: "Err ping",
			st:   &storage.DBStorage{},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/ping/", PingDatabase(tt.st))
			req, _ := http.NewRequest("GET", "/ping/", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestRequestAllMetrics(t *testing.T) {
	tests := []struct {
		st   storage.IStorage
		name string
		code int
	}{
		{
			name: "Normal working",
			st:   &mockStorage{},
			code: 200,
		},
		{
			name: "normal err",
			st:   &storage.DBStorage{},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/", RequestAllMetrics(tt.st))
			req, _ := http.NewRequest("GET", "/", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
			if tt.name != "normal err" {
				assert.NotEqual(t, w.Body.Bytes(), nil)
			}
		})
	}
}

func TestBatchUpdateJSON(t *testing.T) {
	var v float64
	var d int64
	tests := []struct {
		name    string
		st      *mockStorage
		fs      *storage.FileStorage
		key     string
		request []storage.Metrics
		code    int
	}{
		{
			name: "Simple basic test",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "",
			request: []storage.Metrics{
				{ID: "Alloc",
					MType: "gauge",
					Value: &v},
				{ID: "Heap",
					MType: "counter",
					Delta: &d},
			},
			code: 200,
		},
		{
			name: "Simple basic test with key",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			key:  "testkey",

			request: []storage.Metrics{
				{ID: "Alloc",
					MType: "gauge",
					Value: &v,
					Hash:  "9764fd9e51f33bca83ea4218359b3e13257fac2c2ef33dfcb07e68625cfe02e5"},
				{ID: "Heap",
					MType: "counter",
					Delta: &d,
					Hash:  "9764fd9e51f33bca83ea4218359b3e13257fac2c2ef33dfcb07e68625cfe02e5"},
			},
			code: 200,
		},
		{
			name: "Sync test",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			key: "",
			request: []storage.Metrics{
				{ID: "Alloc",
					MType: "gauge",
					Value: &v},
				{ID: "Heap",
					MType: "counter",
					Delta: &d},
			},
			code: 200,
		},
		{
			name: "Sync test",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "",
				Synchronize: true,
			},
			key: "",
			request: []storage.Metrics{
				{ID: "Alloc",
					MType: "gauge",
					Value: &v},
				{ID: "Heap",
					MType: "counter",
					Delta: &d},
			},
			code: 500,
		},
		{
			name: "Insert err",
			st:   &mockStorage{},
			fs:   &storage.FileStorage{},
			request: []storage.Metrics{
				{ID:    "ERROR",
				MType: "counter",
				Delta: &d,},
			},
			code: 500,
		},
		{
			name: "reader err",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			key: "",
			request: []storage.Metrics{
				{ID: "Alloc",
					MType: "gauge",
					Value: &v},
			},
			code: 500,
		},
		{
			name: "unmarshall err",
			st:   &mockStorage{},
			fs: &storage.FileStorage{
				FilePath:    "test",
				Synchronize: true,
			},
			key: "",
			request: []storage.Metrics{
				{ID: "Alloc",
					MType: "gauge",
					Value: &v},
			},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/updates/", BatchUpdateJSON(tt.st, tt.fs, tt.key))
			breq, err := json.Marshal(tt.request)
			if err != nil {
				t.Error(err)
			}
			switch tt.name {
			case "reader err":
				req, _ := http.NewRequest("POST", "/updates/", errReader(0))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
			case "unmarshall err":
				req, _ := http.NewRequest("POST", "/updates/", bytes.NewBuffer([]byte{255, 1, 1, 2, 0, 0}))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
			default:
				req, _ := http.NewRequest("POST", "/updates/", bytes.NewBuffer(breq))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.code, w.Code)
				if tt.fs.Synchronize && tt.fs.FilePath != "" {
					data, err := ioutil.ReadFile(tt.fs.FilePath)
					if err != nil {
						t.Error(err, tt)
					}
					err = os.Remove(tt.fs.FilePath)
					if err != nil {
						t.Error(err, tt)
					}
					assert.Equal(t, string(data), "mock_test")
				}
			}
		})
	}
}

func TestWithoutID(t *testing.T) {
	tests := []struct {
		name string
		url  string
		code int
	}{
		{
			name: "Simple basic test 1",
			url:  "/update/counter/",
			code: 404,
		},
		{
			name: "Simple basic test 2",
			url:  "/update/gauge/",
			code: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/update/gauge/", WithoutID)
			r.POST("/update/counter/", WithoutID)
			req, _ := http.NewRequest("POST", tt.url, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}
