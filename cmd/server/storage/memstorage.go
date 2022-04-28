package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MemoryStorage struct {
	GaugeMetrics   map[string]float64 // хранилище для gauge
	CounterMetrics map[string]int64   // хранилище для counter
	mutex          sync.Mutex
}

func (m *MemoryStorage) UpdateMetricsFromString(metricType, metricName, metricValue string) (int, error) {
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

func (m *MemoryStorage) ReadOldMetrics(path string) error {
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

func (m *MemoryStorage) WriteMetricsToFile(file *os.File) error {
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
