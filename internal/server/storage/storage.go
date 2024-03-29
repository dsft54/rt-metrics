package storage

import (
	"os"
)

// Metrics преобразуемая в json структура, которая может содержать
// тип метрики, её название, значение и хеш
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// IStorage интерфейс описывающий хранище метрик и методы для работы с ним.
type IStorage interface {
	// Write/Insert methods.
	InsertMetric(*Metrics) error
	InsertBatchMetric([]Metrics) error
	ParamsUpdate(string, string, string) (int, error)

	// Read methods.
	ReadMetric(*Metrics) (*Metrics, error)
	ReadAllMetrics() ([]Metrics, error)
	
	// File storage methods.
	SaveToFile(*os.File) error
	UploadFromFile(string) error

	// DB healthcheck.
	Ping() error
}
