package storage

import (
	"context"
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
			name:    "db err gauge",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			m:       &Metrics{
				ID: "Alloc",
				MType: "gauge",
				Value: &v,
			},
			wantErr: true,
		},
		{
			name:    "db err counter",
			d:       &DBStorage{Connection: &pgx.Conn{}},
			m:       &Metrics{
				ID: "Pollcount",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
