package storage

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx"
)

func TestDBStorage_InsertMetric(t *testing.T) {
	var (
		v float64
		d int64
	)
	tests := []struct {
		name    string
		d       *DBStorage
		m       *Metrics
		wantErr bool
	}{
		{
			name:    "connection nil",
			d:       &DBStorage{},
			m:       &Metrics{},
			wantErr: true,
		},
		{
			name: "db err gauge",
			d:    &DBStorage{Connection: &pgx.Conn{}},
			m: &Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: &v,
			},
			wantErr: true,
		},
		{
			name: "db err counter",
			d:    &DBStorage{Connection: &pgx.Conn{}},
			m: &Metrics{
				ID:    "Pollcount",
				MType: "counter",
				Delta: &d,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			if err := tt.d.InsertMetric(tt.m); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.InsertMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_ReadAllMetrics(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		want    []Metrics
		wantErr bool
	}{
		{
			name:    "connection nil",
			d:       &DBStorage{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "connection err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			got, err := tt.d.ReadAllMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.ReadAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBStorage.ReadAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_ParamsUpdate(t *testing.T) {
	tests := []struct {
		name        string
		d           *DBStorage
		metricType  string
		metricID    string
		metricValue string
		want        int
		wantErr     bool
	}{
		{
			name: "no type",
			d: &DBStorage{
				Connection: &pgx.Conn{},
			},
			want:    501,
			wantErr: true,
		},
		{
			name:    "connection nil",
			d:       &DBStorage{},
			want:    500,
			wantErr: true,
		},
		{
			name:        "gauge type parse err",
			metricType:  "gauge",
			metricID:    "Alloc",
			metricValue: "ab",
			d:           &DBStorage{Connection: &pgx.Conn{}},
			want:        400,
			wantErr:     true,
		},
		{
			name:        "gauge type db err",
			metricType:  "gauge",
			metricID:    "Alloc",
			metricValue: "3.14",
			d:           &DBStorage{Connection: &pgx.Conn{}},
			want:        400,
			wantErr:     true,
		},
		{
			name:        "counter type parse err",
			metricType:  "counter",
			metricID:    "Alloc",
			metricValue: "3.14",
			d:           &DBStorage{Connection: &pgx.Conn{}},
			want:        400,
			wantErr:     true,
		},
		{
			name:        "counter type db err",
			metricType:  "counter",
			metricID:    "Alloc",
			metricValue: "3",
			d:           &DBStorage{Connection: &pgx.Conn{}},
			want:        400,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			got, err := tt.d.ParamsUpdate(tt.metricType, tt.metricID, tt.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.ParamsUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DBStorage.ParamsUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_SaveToFile(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		f       *os.File
		wantErr bool
	}{
		{
			name:    "db err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			f:       &os.File{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			if err := tt.d.SaveToFile(tt.f); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_UploadFromFile(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		path    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty path",
			d:       &DBStorage{},
			path:    "",
			wantErr: true,
		},
		{
			name:    "unmarshall err",
			d:       &DBStorage{},
			path:    "teste",
			data:    []byte{255, 0, 0, 0, 1, 223, 12, 5},
			wantErr: true,
		},
		{
			name:    "db gauge err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			path:    "test",
			data:    []byte("[{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":3.14},{\"id\":\"Counter\",\"type\":\"counter\",\"delta\":3}]"),
			wantErr: true,
		},
		{
			name:    "db counter err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			path:    "test",
			data:    []byte("[{\"id\":\"Counter\",\"type\":\"counter\",\"delta\":3},{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":3.14}]"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := os.OpenFile("test", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				t.Error(err)
			}
			_, err = tf.Write(tt.data)
			if err != nil {
				t.Error(err)
			}
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			if err := tt.d.UploadFromFile(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.UploadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			err = tf.Close()
			if err != nil {
				t.Error(err)
			}
		})
	}
	err := os.Remove("test")
	if err != nil {
		t.Error(err)
	}
}

func TestDBStorage_ReadMetric(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		rm      *Metrics
		want    *Metrics
		wantErr bool
	}{
		{
			name:    "db err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			rm:      &Metrics{},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			got, err := tt.d.ReadMetric(tt.rm)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.ReadMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBStorage.ReadMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBStorage_InsertBatchMetric(t *testing.T) {
	var v float64
	tests := []struct {
		name    string
		d       *DBStorage
		metrics []Metrics
		wantErr bool
	}{
		{
			name: "db err",
			d:    &DBStorage{Connection: &pgx.Conn{}},
			metrics: []Metrics{
				{
					ID:    "Alloc",
					MType: "gauge",
					Value: &v,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			if err := tt.d.InsertBatchMetric(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.InsertBatchMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_Ping(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		wantErr bool
	}{
		{
			name:    "db err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			if err := tt.d.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_DBConnectStorage(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		ctx     context.Context
		auth    string
		wantErr bool
	}{
		{
			name:    "dsn parse err",
			d:       &DBStorage{},
			ctx:     context.Background(),
			auth:    "255 0  0 1",
			wantErr: true,
		},
		{
			name:    "connect err",
			d:       &DBStorage{},
			ctx:     context.Background(),
			auth:    "postgres://test:test@localhost:12221",
			wantErr: true,
		},
		{
			name:    "empty auth",
			d:       &DBStorage{},
			ctx:     context.Background(),
			auth:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.DBConnectStorage(tt.ctx, tt.auth); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.DBConnectStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_DBFlushTable(t *testing.T) {
	tests := []struct {
		name    string
		d       *DBStorage
		wantErr bool
	}{
		{
			name:    "db err",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeContext, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			tt.d.Context = timeContext
			if err := tt.d.DBFlushTable(); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.DBFlushTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
