package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dsft54/rt-metrics/cmd/agent/storage"
	"github.com/go-resty/resty/v2"
)

const (
	pollInterval   time.Duration = 2 * time.Second
	reportInterval time.Duration = 10 * time.Second
)

func sendData(url string, m *storage.Metrics) error {
	rawData, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}
	client := resty.New()
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(rawData).
		Post(url)
	return err
}

func reportMetrics(s *storage.Storage, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	reportTicker := time.NewTicker(reportInterval)
	for {
		select {
		case <-reportTicker.C:
			metricsSlice := s.RebuildDataToJSON()
			for _, value := range metricsSlice {
				select {
				case <-ctx.Done():
					return
				default:
					err := sendData("http://localhost:8080/update", &value)
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

func pollMetrics(s *storage.Storage, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	pollTicker := time.NewTicker(pollInterval)
	for {
		select {
		case <-pollTicker.C:
			s.CollectMemMetrics()
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	s := storage.Storage{}
	wg := new(sync.WaitGroup)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(2)
	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go pollMetrics(&s, wg, ctx)
	go reportMetrics(&s, wg, ctx)
	sig := <-syscallCancelChan
	cancel()
	fmt.Printf("Caught syscall: %v", sig)
	wg.Wait()
}
