package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dsft54/rt-metrics/cmd/server/settings"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MetricStorages struct {
	GaugeMetrics   map[string]float64 // хранилище для gauge
	CounterMetrics map[string]int64   // хранилище для counter
	mutex          sync.Mutex
}

func (m *MetricStorages) UpdateMetricsFromString(metricType, metricName, metricValue string) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	switch metricType {
	case "gauge":
		floatFromString, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return 400, err
		}
		m.GaugeMetrics[metricName] = floatFromString
		return 200, nil
	case "counter":
		intFromString, err := strconv.Atoi(metricValue)
		if err != nil {
			return 400, err
		}
		m.CounterMetrics[metricName] += int64(intFromString)
		return 200, nil
	default:
		return 501, errors.New("wrong metric type - " + metricType)
	}
}

func (m *MetricStorages) ReadOldMetrics(path string) error {
	var metricsSlice []Metrics

	m.mutex.Lock()
	defer m.mutex.Unlock()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &metricsSlice)
	if err != nil {
		return err
	}

	for _, val := range metricsSlice {
		switch val.MType {
		case "gauge":
			m.GaugeMetrics[val.ID] = *val.Value
		case "counter":
			m.CounterMetrics[val.ID] += *val.Delta
		}
	}
	return nil
}

func (m *MetricStorages) WriteMetricsToFile(file *os.File) error {
	var metricsSlice []Metrics

	m.mutex.Lock()
	for key, value := range m.GaugeMetrics {
		v := value
		metricsSlice = append(metricsSlice, Metrics{
			ID:    key,
			MType: "gauge",
			Value: &v,
		})
	}
	for key, value := range m.CounterMetrics {
		v := value
		metricsSlice = append(metricsSlice, Metrics{
			ID:    key,
			MType: "counter",
			Delta: &v,
		})
	}
	m.mutex.Unlock()

	data, err := json.Marshal(metricsSlice)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

type FileStorage struct {
	File        *os.File
	FilePath    string
	StoreData   bool
	Synchronize bool
}

func (f *FileStorage) OpenToWrite(path string) error {
	var err error
	f.File, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) SaveDataToFile(condition bool, m *MetricStorages) error {
	if condition {
		err := f.OpenToWrite(f.FilePath)
		if err != nil {
			return err
		}
		m.WriteMetricsToFile(f.File)
		f.File.Close()
	}
	return nil
}

func (f *FileStorage) InitFileStorage(cfg settings.Config) {
	if cfg.StoreFile == "" {
		f.StoreData = false
		f.Synchronize = false
	} else {
		f.FilePath = cfg.StoreFile
		f.StoreData = true
	}
	if cfg.StoreInterval == 0 {
		f.Synchronize = true
	}
	if cfg.StoreInterval > 0 {
		f.Synchronize = false
	}
}

func (f *FileStorage) IntervalUpdate(ctx context.Context, dur time.Duration, s *MetricStorages) {
	intervalTicker := time.NewTicker(dur)
	for {
		select {
		case <-intervalTicker.C:
			f.OpenToWrite(f.FilePath)
			s.WriteMetricsToFile(f.File)
			f.File.Close()
		case <-ctx.Done():
			return
		}
	}
}
