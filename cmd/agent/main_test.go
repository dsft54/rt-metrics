package main

import (
	"testing"
)

func Test_sendData(t *testing.T) {
	type args struct {
		url   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "server offline (or not correct)",
			args: args{
				url:   "http://localhost:808",
				value: "/update/gauge/Alloc/3.1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendData(tt.args.url, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("sendData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rebuildData(t *testing.T) {
	tests := []struct {
		name    string
		metrics *Metric
		want    string
	}{
		{
			name: "Normal conditions",
			metrics: &Metric{
				Alloc: 3.14159265,
			},
			want: "/gauge/Alloc/3.14159265",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rebuildData(tt.metrics)
			if got[0] != tt.want {
				t.Errorf("rebuildData() = %v, want %v", got[0], tt.want)
			}
		})
	}
}

func TestMetric_collectMemMetrics(t *testing.T) {
	tests := []struct {
		name string
		m    *Metric
	}{
		{
			name: "Normal conditions",
			m:    &Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.collectMemMetrics()
			if tt.m.PollCount != 1 {
				t.Errorf("PollCount is still zero")
			}
		})
	}
}
