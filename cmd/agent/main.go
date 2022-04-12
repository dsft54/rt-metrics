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

	"github.com/caarlos0/env"
	"github.com/dsft54/rt-metrics/cmd/agent/storage"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
}

func sendData(url string, m *storage.Metrics, client *resty.Client) error {
	rawData, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(rawData).
		Post(url)
	return err
}

func reportMetrics(addr string, interval time.Duration, s *storage.Storage, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	client := resty.New()
	reportTicker := time.NewTicker(interval)
	for {
		select {
		case <-reportTicker.C:
			metricsSlice := s.RebuildDataToJSON()
			for _, value := range metricsSlice {
				select {
				case <-ctx.Done():
					return
				default:
					err := sendData("http://"+addr+"/update", &value, client)
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

func pollMetrics(interval time.Duration, s *storage.Storage, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	pollTicker := time.NewTicker(interval)
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
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	s := storage.Storage{}
	wg := new(sync.WaitGroup)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(2)
	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go pollMetrics(cfg.PollInterval, &s, wg, ctx)
	go reportMetrics(cfg.Address, cfg.ReportInterval, &s, wg, ctx)
	sig := <-syscallCancelChan
	cancel()
	fmt.Printf("Caught syscall: %v", sig)
	wg.Wait()
}
