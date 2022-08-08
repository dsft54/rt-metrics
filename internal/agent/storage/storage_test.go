// Модуль storage определяет структуры их методы, предназначенные для описания хранилища текущего значения метрик,

// из которого они будут отправлены на сервер.

package storage

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func ExampleMemStorage_ConvertToURLParams() {
	ms := &MemStorage{
		GaugeMetrics:   map[string]gauge{"Alloc": 3.14},
		CounterMetrics: map[string]counter{},
	}
	out1 := ms.ConvertToURLParams()
	sort.Strings(out1)
	fmt.Println(out1)

	ms = &MemStorage{
		GaugeMetrics:   map[string]gauge{"Alloc": 3.14, "Heap": 6.28},
		CounterMetrics: map[string]counter{"Counter": 1},
	}
	out2 := ms.ConvertToURLParams()
	sort.Strings(out2)

	fmt.Println(out2)

	// Output:
	// [/gauge/Alloc/3.14]
	// [/counter/Counter/1 /gauge/Alloc/3.14 /gauge/Heap/6.28]
}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "constructor test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMemStorage()
			if got.CounterMetrics == nil || got.GaugeMetrics == nil {
				t.Error("New memory storage cannot be created")
			}
		})
	}
}

func TestMemStorage_CollectRuntimeMetrics(t *testing.T) {
	tests := []struct {
		ms   *MemStorage
		name string
	}{
		{
			name: "normal conditions test",
			ms:   NewMemStorage(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ms.CollectRuntimeMetrics()
			if tt.ms.GaugeMetrics["Alloc"] == 0 || tt.ms.CounterMetrics["PollCount"] != 1 {
				t.Error("Cannot collect runtime metrics")
			}
		})
	}
}

func TestMemStorage_CollectPSUtilMetrics(t *testing.T) {
	tests := []struct {
		ms      *MemStorage
		name    string
		wantErr bool
	}{
		{
			name:    "normal conditions test",
			ms:      NewMemStorage(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ms.CollectPSUtilMetrics(); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.CollectPSUtilMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.ms.GaugeMetrics["TotalMemory"] == 0 {
				t.Error("Cannot collect PSutil metrics")
			}
		})
	}
}

func TestMemStorage_ConvertToMetricsJSON(t *testing.T) {
	var testDelta int64
	testDelta, testValue := 3, 3.14
	tests := []struct {
		name string
		ms   *MemStorage
		hkey string
		want []Metrics
	}{
		{
			name: "normal conditions test",
			ms: &MemStorage{
				GaugeMetrics: map[string]gauge{
					"Alloc": 3.14,
				},
				CounterMetrics: map[string]counter{
					"PollCounter": 3,
				},
			},
			hkey: "",
			want: []Metrics{
				{
					ID: "Alloc",
					MType: "gauge",
					Value: &testValue,
				},
				{
					ID: "PollCounter",
					MType: "counter",
					Delta: &testDelta,
				},
			},
		},
		{
			name: "normal conditions test with hash key",
			ms: &MemStorage{
				GaugeMetrics: map[string]gauge{
					"Alloc": 3.14,
				},
				CounterMetrics: map[string]counter{
					"PollCounter": 3,
				},
			},
			hkey: "test",
			want: []Metrics{
				{
					ID: "Alloc",
					MType: "gauge",
					Value: &testValue,
					Hash: "a1d545936b9abb1b47000a9a5717b367544cbdd20e796c8d70d9ec2785cf9629",
				},
				{
					ID: "PollCounter",
					MType: "counter",
					Delta: &testDelta,
					Hash: "8986ad2b836333c6845884d1512181faaa7587854a54b681cbff6ebd25a7d041",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ms.ConvertToMetricsJSON(tt.hkey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.ConvertToMetricsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
