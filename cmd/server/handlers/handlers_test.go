package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/dsft54/rt-metrics/cmd/server/storage"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert"
)

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

func (ms *mockStorage) InsertMetric(*storage.Metrics) error {
	return nil
}

func (ms *mockStorage) InsertBatchMetric(metrics []storage.Metrics) error {
	return nil
}

func (ms *mockStorage) ReadMetric(mr *storage.Metrics) (*storage.Metrics, error) {
	var value float64
	if mr.MType != "gauge" && mr.ID != "Alloc" {
		return nil, errNotFound
	}
	mr.Value = &value
	return mr, nil
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
			name: "Existing metric test",
			st:   &mockStorage{},
			url:  "/value/gauge/Alloc",
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
	type args struct {
		st  storage.Storage
		key string
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequestMetricJSON(tt.args.st, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RequestMetricJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
