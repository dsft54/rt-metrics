package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type (
	gauge   float64
	counter int64
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MemStorage struct {
	sync.RWMutex
	GaugeMetrics   map[string]gauge
	CounterMetrics map[string]counter
}

func NewMemStorage() *MemStorage {
	ms := MemStorage{
		GaugeMetrics:   make(map[string]gauge),
		CounterMetrics: make(map[string]counter),
	}
	return &ms
}

var memstats runtime.MemStats

func (ms *MemStorage) CollectRuntimeMetrics() {
	runtime.ReadMemStats(&memstats)
	ms.Lock()
	defer ms.Unlock()
	ms.GaugeMetrics["Alloc"] = gauge(memstats.Alloc)
	ms.GaugeMetrics["BuckHashSys"] = gauge(memstats.BuckHashSys)
	ms.GaugeMetrics["Frees"] = gauge(memstats.Frees)
	ms.GaugeMetrics["GCCPUFraction"] = gauge(memstats.GCCPUFraction)
	ms.GaugeMetrics["GCSys"] = gauge(memstats.GCSys)
	ms.GaugeMetrics["HeapAlloc"] = gauge(memstats.HeapAlloc)
	ms.GaugeMetrics["HeapIdle"] = gauge(memstats.HeapIdle)
	ms.GaugeMetrics["HeapInuse"] = gauge(memstats.HeapInuse)
	ms.GaugeMetrics["HeapObjects"] = gauge(memstats.HeapObjects)
	ms.GaugeMetrics["HeapReleased"] = gauge(memstats.HeapReleased)
	ms.GaugeMetrics["HeapSys"] = gauge(memstats.HeapSys)
	ms.GaugeMetrics["LastGC"] = gauge(memstats.LastGC)
	ms.GaugeMetrics["Lookups"] = gauge(memstats.Lookups)
	ms.GaugeMetrics["MCacheInuse"] = gauge(memstats.MCacheInuse)
	ms.GaugeMetrics["MCacheSys"] = gauge(memstats.MCacheSys)
	ms.GaugeMetrics["MSpanInuse"] = gauge(memstats.MSpanInuse)
	ms.GaugeMetrics["MSpanSys"] = gauge(memstats.MSpanSys)
	ms.GaugeMetrics["Mallocs"] = gauge(memstats.Mallocs)
	ms.GaugeMetrics["NextGC"] = gauge(memstats.NextGC)
	ms.GaugeMetrics["NumForcedGC"] = gauge(memstats.NumForcedGC)
	ms.GaugeMetrics["NumGC"] = gauge(memstats.NumGC)
	ms.GaugeMetrics["OtherSys"] = gauge(memstats.OtherSys)
	ms.GaugeMetrics["PauseTotalNs"] = gauge(memstats.PauseTotalNs)
	ms.GaugeMetrics["StackInuse"] = gauge(memstats.StackInuse)
	ms.GaugeMetrics["StackSys"] = gauge(memstats.StackSys)
	ms.GaugeMetrics["Sys"] = gauge(memstats.Sys)
	ms.GaugeMetrics["TotalAlloc"] = gauge(memstats.TotalAlloc)
	ms.GaugeMetrics["RandomValue"] = gauge(rand.Float64())
	ms.CounterMetrics["PollCount"] += 1
}

func (ms *MemStorage) CollectPSUtilMetrics() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	c, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}
	ms.Lock()
	ms.GaugeMetrics["TotalMemory"] = gauge(v.Total)
	ms.GaugeMetrics["FreeMemory"] = gauge(v.Free)
	for i, value := range c {
		ms.GaugeMetrics["CPUutilization"+strconv.Itoa(i+1)] = gauge(value)
	}
	ms.Unlock()
	return nil
}

func (ms *MemStorage) ConvertToMetricsJSON(hkey string) []Metrics {
	metricsSlice := []Metrics{}
	ms.RLock()
	for id, value := range ms.GaugeMetrics {
		metricsPart := Metrics{MType: "gauge", ID: id}
		v := float64(value)
		metricsPart.Value = &v
		if hkey != "" {
			h := hmac.New(sha256.New, []byte(hkey))
			h.Write([]byte(fmt.Sprintf("%s:gauge:%f", id, v)))
			metricsPart.Hash = hex.EncodeToString(h.Sum(nil))
		}
		metricsSlice = append(metricsSlice, metricsPart)
	}
	for id, delta := range ms.CounterMetrics {
		metricsPart := Metrics{MType: "counter", ID: id}
		v := int64(delta)
		metricsPart.Delta = &v
		if hkey != "" {
			h := hmac.New(sha256.New, []byte(hkey))
			h.Write([]byte(fmt.Sprintf("%s:counter:%d", id, v)))
			metricsPart.Hash = hex.EncodeToString(h.Sum(nil))
		}
		metricsSlice = append(metricsSlice, metricsPart)
	}
	ms.RUnlock()
	return metricsSlice
}

func (ms *MemStorage) ConvertToURLParams() []string {
	urlsList := []string{}
	ms.RLock()
	for id, value := range ms.GaugeMetrics {
		urlsList = append(urlsList, fmt.Sprintf("/%s/%s/%v", "gauge", id, value))
	}
	for id, delta := range ms.CounterMetrics {
		urlsList = append(urlsList, fmt.Sprintf("/%s/%s/%v", "counter", id, delta))
	}
	ms.RUnlock()
	return urlsList
}
