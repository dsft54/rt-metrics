package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/dsft54/rt-metrics/cmd/server/settings"
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

func (m *MetricStorages) ReadOldMetrics() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var metricsSlice []Metrics

	data, err := ioutil.ReadFile(settings.Cfg.StoreFile)
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
			Store.GaugeMetrics[val.ID] = *val.Value
		case "counter":
			Store.CounterMetrics[val.ID] = *val.Delta
		}
	}
	return nil
}

func (m *MetricStorages) WriteMetricsToFile(file *os.File) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var metricsSlice []Metrics
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
	data, err := json.MarshalIndent(metricsSlice, "", "")
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
	Synchronize bool
}

func (f *FileStorage) OpenToWrite() error {
	var err error
	f.File, err = os.OpenFile(settings.Cfg.StoreFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) InitFileStorage(cfg settings.Config) error {
	if cfg.Restore {
		Store.ReadOldMetrics()
	}
	if cfg.StoreInterval == 0 {
		f.Synchronize = true
	}
	if cfg.StoreInterval > 0 {
		f.Synchronize = false
	}
	return nil
}

func (f *FileStorage) IntervalUpdate(dur time.Duration) {
	intervalTicker := time.NewTicker(dur)
	for {
		<-intervalTicker.C
		f.OpenToWrite()
		Store.WriteMetricsToFile(f.File)
		f.File.Close()
	}
}

var (
	Store     MetricStorages
	FileStore FileStorage
)
