package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	pollInterval   time.Duration = 2 * time.Second
	reportInterval time.Duration = 4 * time.Second
)

type (
	gauge   float64
	counter int64
)

type Metric struct {
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
	Mutex         sync.Mutex
}

var memstats runtime.MemStats

func (m *Metric) collectMemMetrics() {
	runtime.ReadMemStats(&memstats)
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Alloc = gauge(memstats.Alloc)
	m.BuckHashSys = gauge(memstats.BuckHashSys)
	m.Frees = gauge(memstats.Frees)
	m.GCCPUFraction = gauge(memstats.GCCPUFraction)
	m.GCSys = gauge(memstats.GCSys)
	m.HeapAlloc = gauge(memstats.HeapAlloc)
	m.HeapIdle = gauge(memstats.HeapIdle)
	m.HeapInuse = gauge(memstats.HeapInuse)
	m.HeapObjects = gauge(memstats.HeapObjects)
	m.HeapReleased = gauge(memstats.HeapReleased)
	m.HeapSys = gauge(memstats.HeapSys)
	m.LastGC = gauge(memstats.LastGC)
	m.Lookups = gauge(memstats.Lookups)
	m.MCacheInuse = gauge(memstats.MCacheInuse)
	m.MCacheSys = gauge(memstats.MCacheSys)
	m.MSpanInuse = gauge(memstats.MSpanInuse)
	m.MSpanSys = gauge(memstats.MSpanSys)
	m.Mallocs = gauge(memstats.Mallocs)
	m.NextGC = gauge(memstats.NextGC)
	m.NumForcedGC = gauge(memstats.NumForcedGC)
	m.NumGC = gauge(memstats.NumGC)
	m.OtherSys = gauge(memstats.OtherSys)
	m.PauseTotalNs = gauge(memstats.PauseTotalNs)
	m.StackInuse = gauge(memstats.StackInuse)
	m.StackSys = gauge(memstats.StackSys)
	m.Sys = gauge(memstats.Sys)
	m.TotalAlloc = gauge(memstats.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = gauge(rand.Float64())
}

func rebuildData(m *Metric) []string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	value := reflect.ValueOf(m).Elem()
	typeOfS := value.Type()
	usrlList := []string{}
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Type().Name() == "Mutex" {
			continue
		}
		usrlList = append(usrlList, fmt.Sprintf("/%s/%s/%v", value.Field(i).Type().Name(), typeOfS.Field(i).Name, value.Field(i).Interface()))
	}
	return usrlList
}

func sendData(url, value string) error {
	client := resty.New()
	_, err := client.R().Post(url + value)
	return err
}

func pollMetrics(m *Metric, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	pollTicker := time.NewTicker(pollInterval)
	for {
		select {
		case <-pollTicker.C:
			m.collectMemMetrics()
		case <-ctx.Done():
			return
		}
	}
}

func reportMetrics(m *Metric, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	reportTicker := time.NewTicker(reportInterval)
	for {
		select {
		case <-reportTicker.C:
			valuesURLs := rebuildData(m)
			for _, value := range valuesURLs {
				select {
				case <-ctx.Done():
					return
				default:
					err := sendData("http://localhost:8080/update", value)
					if err != nil {
						continue
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	m := Metric{}
	wg := new(sync.WaitGroup)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(2)
	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go pollMetrics(&m, wg, ctx)
	go reportMetrics(&m, wg, ctx)
	sig := <-syscallCancelChan
	cancel()
	fmt.Printf("Caught syscall: %v", sig)
	wg.Wait()
}
