package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-resty/resty/v2"
	
	"github.com/dsft54/rt-metrics/cmd/agent/settings"
	"github.com/dsft54/rt-metrics/cmd/agent/storage"
)

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

func reportMetrics(ctx context.Context, cfg *settings.Config, s *storage.Storage, wg *sync.WaitGroup) {
	defer wg.Done()
	client := resty.New()
	reportTicker := time.NewTicker(cfg.ReportInterval)
	for {
		select {
		case <-reportTicker.C:
			metricsSlice := s.RebuildDataToJSON(cfg.HashKey)
			for _, value := range metricsSlice {
				select {
				case <-ctx.Done():
					return
				default:
					err := sendData("http://"+cfg.Address+"/update", &value, client)
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

func pollMetrics(ctx context.Context, interval time.Duration, s *storage.Storage, wg *sync.WaitGroup) {
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

func init() {
	flag.StringVar(&config.Address, "a", "localhost:8080", "Metrics server address")
	flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "Runtime poll interval")
	flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "Report metrics interval")
	flag.StringVar(&config.HashKey, "k", "", "SHA256 signing key")
}

var config settings.Config

func main() {
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}
	s := storage.Storage{}
	wg := new(sync.WaitGroup)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(2)

	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go pollMetrics(ctx, config.PollInterval, &s, wg)
	go reportMetrics(ctx, &config, &s, wg)
	sig := <-syscallCancelChan
	cancel()
	fmt.Printf("Caught syscall: %v", sig)
	wg.Wait()
}
