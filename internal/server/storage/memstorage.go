package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

var (
	errWrType = fmt.Errorf("insert data: wrong type")
	errNoDB     = fmt.Errorf("no db connected")
	errNotFound = fmt.Errorf("not found in memory storage")
)

// MemoryStorage структура, состоящая из двух массивов для двух типов метрик gauge и counter,
// а также RWMutex для конкурентного доступа к ним. Реализует интерфейсный тип IStorage.
type MemoryStorage struct {
	GaugeMetrics   map[string]float64 // хранилище для gauge
	CounterMetrics map[string]int64   // хранилище для counter
	mutex          sync.RWMutex
}

// InsertMetric потокобезопасно добавляет значения полученные из аргумента Metrics в массивы.
// Для gauge заменяет существующий, для counter добавляет к уже существующему значению в базе.
func (m *MemoryStorage) InsertMetric(met *Metrics) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	switch met.MType {
	case "gauge":
		m.GaugeMetrics[met.ID] = *met.Value
	case "counter":
		m.CounterMetrics[met.ID] += *met.Delta
	default:
		return errWrType
	}
	return nil
}

// InsertBatchMetric в цикле выполняет InsertMetric из значений, полученных в качестве аргумента
// в виде списка []Metrics.
func (m *MemoryStorage) InsertBatchMetric(metrics []Metrics) error {
	for _, metric := range metrics {
		err := m.InsertMetric(&metric)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadMetric потокобезопасно получает значение искомой метрики и возвращает в виде структуры Metrics.
// В случае если искомой метрики не существует, возвращает nil и ошибку.
func (m *MemoryStorage) ReadMetric(rm *Metrics) (*Metrics, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	switch rm.MType {
	case "gauge":
		if value, found := m.GaugeMetrics[rm.ID]; found {
			rm.Value = &value
		} else {
			return nil, errNotFound
		}
	case "counter":
		if value, found := m.CounterMetrics[rm.ID]; found {
			rm.Delta = &value
		} else {
			return nil, errNotFound
		}
	}
	return rm, nil
}

// ReadAllMetrics потокобезопасно получает значение всех метрик и возвращает в виде списка структур []Metrics.
// В случае если метрик неn, возвращает пустой список и nil.
func (m *MemoryStorage) ReadAllMetrics() ([]Metrics, error) {
	metricsSlice := []Metrics{}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for key, value := range m.GaugeMetrics {
		metric := Metrics{
			MType: "gauge",
			ID:    key,
			Value: &value,
		}
		metricsSlice = append(metricsSlice, metric)
	}
	for key, value := range m.CounterMetrics {
		metric := Metrics{
			MType: "counter",
			ID:    key,
			Delta: &value,
		}
		metricsSlice = append(metricsSlice, metric)
	}
	return metricsSlice, nil
}

// ParamsUpdate потокобезопасно добавляет значения полученные из строчных аргументов в массивы.
// Для gauge заменяет существующий, для counter добавляет к уже существующему значению в базе.
// А также возвращает код, в зависимости от успешности операции для передачи его в handler.
func (m *MemoryStorage) ParamsUpdate(metricType, metricName, metricValue string) (int, error) {
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

// UploadFromFile потокобезопасно заполняет массивы метриками, полученными
// из файла по пути.
func (m *MemoryStorage) UploadFromFile(path string) error {
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

// SaveToFile сохраняет значения метрик в файл, собирая их в список структур Metrics.
func (m *MemoryStorage) SaveToFile(file *os.File) error {
	var metricsSlice []Metrics
	m.mutex.Lock()
	defer m.mutex.Unlock()
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

// Ping всегда возвращает ошибку. Необходим для реализации интерфейса.
func (m *MemoryStorage) Ping() error {
	return errNoDB
}
