package storage

import (
	"sync"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MetricStorages struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
	mutex          sync.Mutex
}

var Store MetricStorages
