package storage

import (
	"os"
	"reflect"
	"testing"

	"github.com/go-playground/assert"
)

func TestMemoryStorage_InsertMetric(t *testing.T) {
	var v float64
	tests := []struct {
		m         *MemoryStorage
		wantMemSt *MemoryStorage
		met       *Metrics
		name      string
		wantErr   bool
	}{
		{
			name: "Basic insert test",
			m: &MemoryStorage{
				GaugeMetrics:   make(map[string]float64),
				CounterMetrics: make(map[string]int64),
			},
			met: &Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: &v,
			},
			wantMemSt: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": v,
				},
				CounterMetrics: map[string]int64{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.InsertMetric(tt.met); (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.InsertMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.m, tt.wantMemSt)
		})
	}
}

func TestMemoryStorage_InsertBatchMetric(t *testing.T) {
	var (
		v float64
		d int64
	)
	tests := []struct {
		m         *MemoryStorage
		wantMemSt *MemoryStorage
		name      string
		metrics   []Metrics
		wantErr   bool
	}{
		{
			name: "Basic insert test",
			m: &MemoryStorage{
				GaugeMetrics:   make(map[string]float64),
				CounterMetrics: make(map[string]int64),
			},
			metrics: []Metrics{
				{
					MType: "gauge",
					ID:    "Alloc",
					Value: &v,
				},
				{
					MType: "counter",
					ID:    "Counter",
					Delta: &d,
				},
			},
			wantMemSt: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": v,
				},
				CounterMetrics: map[string]int64{
					"Counter": d,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.InsertBatchMetric(tt.metrics); (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.InsertMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.m, tt.wantMemSt)
		})
	}
}

func TestMemoryStorage_ReadMetric(t *testing.T) {
	v := 3.14
	tests := []struct {
		m       *MemoryStorage
		rm      *Metrics
		want    *Metrics
		name    string
		wantErr bool
	}{
		{
			name: "Basic read test",
			m: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": v,
				},
				CounterMetrics: map[string]int64{},
			},
			rm: &Metrics{
				MType: "gauge",
				ID:    "Alloc",
			},
			want: &Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: &v,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.ReadMetric(tt.rm)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.ReadMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryStorage.ReadMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStorage_ReadAllMetrics(t *testing.T) {
	var d int64 = 3
	v := 3.14
	tests := []struct {
		name    string
		m       *MemoryStorage
		want    []Metrics
		wantErr bool
	}{
		{
			name: "Basic read test",
			m: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": v,
				},
				CounterMetrics: map[string]int64{
					"Counter": d,
				},
			},
			want: []Metrics{
				{
					MType: "gauge",
					ID:    "Alloc",
					Value: &v,
				},
				{
					MType: "counter",
					ID:    "Counter",
					Delta: &d,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.ReadAllMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.ReadAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryStorage.ReadAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStorage_ParamsUpdate(t *testing.T) {
	tests := []struct {
		m           *MemoryStorage
		wantMemSt   *MemoryStorage
		name        string
		metricType  string
		metricName  string
		metricValue string
		want        int
		wantErr     bool
	}{
		{
			name: "Basic insert test",
			m: &MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			metricType:  "gauge",
			metricName:  "Alloc",
			metricValue: "3.14",
			want:        200,
			wantMemSt: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": 3.14,
				},
				CounterMetrics: map[string]int64{},
			},
			wantErr: false,
		},
		{
			name: "Wrong type test",
			m: &MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			metricType:  "wrong",
			metricName:  "Alloc",
			metricValue: "3.14",
			want:        501,
			wantMemSt: &MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			wantErr: true,
		},
		{
			name: "Conv error test",
			m: &MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			metricType:  "gauge",
			metricName:  "Alloc",
			metricValue: "3.a14",
			want:        400,
			wantMemSt: &MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.ParamsUpdate(tt.metricType, tt.metricName, tt.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.ParamsUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MemoryStorage.ParamsUpdate() = %v, want %v", got, tt.want)
			}
			assert.Equal(t, tt.m, tt.wantMemSt)
		})
	}
}

func TestMemoryStorage_UploadFromFile(t *testing.T) {
	tf, err := os.OpenFile("test", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		t.Error(err)
	}
	_, err = tf.Write([]byte("[{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":3.14},{\"id\":\"Counter\",\"type\":\"counter\",\"delta\":3}]"))
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		m         *MemoryStorage
		wantMemSt *MemoryStorage
		name      string
		path      string
		wantErr   bool
	}{
		{
			name: "Basic upload test",
			m: &MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			path: "test",
			wantMemSt: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": 3.14,
				},
				CounterMetrics: map[string]int64{
					"Counter": 3,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err = tt.m.UploadFromFile(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.UploadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.m, tt.wantMemSt)
		})
	}
	err = tf.Close()
	if err != nil {
		t.Error(err)
	}
	err = os.Remove(tf.Name())
	if err != nil {
		t.Error(err)
	}
}

func TestMemoryStorage_SaveToFile(t *testing.T) {
	tests := []struct {
		name    string
		m       *MemoryStorage
		path    string
		wantErr bool
	}{
		{
			name: "Basic save test",
			m: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": 3.14,
				},
				CounterMetrics: map[string]int64{},
			},
			path:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := os.OpenFile(tt.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				t.Error(err)
			}
			if err := tt.m.SaveToFile(tf); (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.path != "" {
				err := tf.Close()
				if err != nil {
					t.Error(err)
				}
				err = os.Remove(tf.Name())
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestMemoryStorage_Ping(t *testing.T) {
	tests := []struct {
		m       *MemoryStorage
		name    string
		wantErr bool
	}{
		{
			name: "Basic ping test",
			m: &MemoryStorage{
				GaugeMetrics: map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
