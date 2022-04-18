package main

import (
	"testing"
	
	"github.com/go-resty/resty/v2"

	"github.com/dsft54/rt-metrics/cmd/agent/storage"
)

func Test_sendData(t *testing.T) {
	client := resty.New()
	type args struct {
		url     string
		metrics storage.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "server offline (or not correct)",
			args: args{
				url: "http://localhost:808",
				metrics: storage.Metrics{
					ID:    "",
					MType: "",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendData(tt.args.url, &tt.args.metrics, client); (err != nil) != tt.wantErr {
				t.Errorf("sendData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetric_collectMemMetrics(t *testing.T) {
	tests := []struct {
		name string
		m    *storage.Storage
	}{
		{
			name: "Normal conditions",
			m:    &storage.Storage{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.CollectMemMetrics()
			if tt.m.PollCount != 1 {
				t.Errorf("PollCount is still zero")
			}
		})
	}
}
