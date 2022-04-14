package storage

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
)

type (
	gauge   float64
	counter int64
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Storage struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge
	PollCount     counter
	RandomValue   gauge
	mutex         sync.Mutex
}

var memstats runtime.MemStats

func (s *Storage) CollectMemMetrics() {
	runtime.ReadMemStats(&memstats)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Alloc = gauge(memstats.Alloc)
	s.BuckHashSys = gauge(memstats.BuckHashSys)
	s.Frees = gauge(memstats.Frees)
	s.GCCPUFraction = gauge(memstats.GCCPUFraction)
	s.GCSys = gauge(memstats.GCSys)
	s.HeapAlloc = gauge(memstats.HeapAlloc)
	s.HeapIdle = gauge(memstats.HeapIdle)
	s.HeapInuse = gauge(memstats.HeapInuse)
	s.HeapObjects = gauge(memstats.HeapObjects)
	s.HeapReleased = gauge(memstats.HeapReleased)
	s.HeapSys = gauge(memstats.HeapSys)
	s.LastGC = gauge(memstats.LastGC)
	s.Lookups = gauge(memstats.Lookups)
	s.MCacheInuse = gauge(memstats.MCacheInuse)
	s.MCacheSys = gauge(memstats.MCacheSys)
	s.MSpanInuse = gauge(memstats.MSpanInuse)
	s.MSpanSys = gauge(memstats.MSpanSys)
	s.Mallocs = gauge(memstats.Mallocs)
	s.NextGC = gauge(memstats.NextGC)
	s.NumForcedGC = gauge(memstats.NumForcedGC)
	s.NumGC = gauge(memstats.NumGC)
	s.OtherSys = gauge(memstats.OtherSys)
	s.PauseTotalNs = gauge(memstats.PauseTotalNs)
	s.StackInuse = gauge(memstats.StackInuse)
	s.StackSys = gauge(memstats.StackSys)
	s.Sys = gauge(memstats.Sys)
	s.TotalAlloc = gauge(memstats.TotalAlloc)
	s.PollCount += 1
	s.RandomValue = gauge(rand.Float64())
}

func (s *Storage) RebuildDataToJSON() []Metrics {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	metricsSlice := []Metrics{}
	value := reflect.ValueOf(s).Elem()
	typeOfS := value.Type()
	for i := 0; i < value.NumField(); i++ {
		if typeOfS.Field(i).Name == "mutex" {
			continue
		}
		metricsPart := Metrics{}
		if value.Field(i).Type().Name() == "gauge" {
			metricsPart.MType = value.Field(i).Type().Name()
			v := float64(value.Field(i).Interface().(gauge))
			metricsPart.Value = &v
		}
		if value.Field(i).Type().Name() == "counter" {
			metricsPart.MType = value.Field(i).Type().Name()
			v := int64(value.Field(i).Interface().(counter))
			metricsPart.Delta = &v
		}
		metricsPart.ID = typeOfS.Field(i).Name
		metricsSlice = append(metricsSlice, metricsPart)
	}
	return metricsSlice
}

func (s *Storage) RebuildDataToString() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	value := reflect.ValueOf(s).Elem()
	typeOfS := value.Type()
	urlsList := []string{}
	for i := 0; i < value.NumField(); i++ {
		if typeOfS.Field(i).Name == "mutex" {
			continue
		}
		urlsList = append(urlsList, fmt.Sprintf("/%s/%s/%v", value.Field(i).Type().Name(), typeOfS.Field(i).Name, value.Field(i).Interface()))
	}
	return urlsList
}
