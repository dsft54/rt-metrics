package storage

import (
	"os"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type Storage interface {
	InsertMetric(*Metrics) error
	InsertBatchMetric([]Metrics) error
	ParamsUpdate(string, string, string) (int, error)
	ReadMetric(*Metrics) (*Metrics, error)
	ReadAllMetrics() ([]Metrics, error)
	SaveToFile(*os.File) error
	UploadFromFile(string) error
}
