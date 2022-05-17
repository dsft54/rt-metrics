package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
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

func sendData(url string, m interface{}, client *resty.Client) error {
	rawData, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(rawData).
		Post(url)
	return err
}

func reportMetrics(ctx context.Context, cfg *settings.Config, s *storage.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	client := resty.New()
	reportTicker := time.NewTicker(cfg.ReportInterval)
	for {
		select {
		case <-reportTicker.C:
			metricsSlice := s.ConvertToMetricsJSON(cfg.HashKey)
			if !cfg.Batched {
				for _, value := range metricsSlice {
					select {
					case <-ctx.Done():
						return
					default:
						err := sendData("http://"+cfg.Address+"/update", &value, client)
						if err != nil {
							log.Println(err)
							continue
						}
					}
				}
			} else {
				err := sendData("http://"+cfg.Address+"/updates", &metricsSlice, client)
				if err != nil {
					log.Println(err)
				}
			}
			log.Println("Atempted to report all metrics. Interval", cfg.ReportInterval)
			log.Println("Atempted to report all metrics. Data", metricsSlice)
		case <-ctx.Done():
			return
		}
	}
}

func pollScheduller(ctx context.Context, c *sync.Cond, t *time.Ticker, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-t.C:
			c.Broadcast()
		case <-ctx.Done():
			return
		}
	}
}

func pollRuntimeMetrics(ctx context.Context, c *sync.Cond, s *storage.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.L.Lock()
			s.CollectRuntimeMetrics()
			log.Println("All runtime memory stats collected")
			c.Wait()
			c.L.Unlock()
		}
	}
}

func pollPSUtilMetrics(ctx context.Context, c *sync.Cond, s *storage.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.L.Lock()
			s.CollectPSUtilMetrics()
			log.Println("All psutil memory stats collected")
			c.Wait()
			c.L.Unlock()
		}
	}
}

func init() {
	flag.StringVar(&config.Address, "a", "localhost:8080", "Metrics server address")
	flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "Runtime poll interval")
	flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "Report metrics interval")
	flag.BoolVar(&config.Batched, "b", true, "Batched metric report")
	flag.StringVar(&config.HashKey, "k", "", "SHA256 signing key")
}

var config settings.Config

func main() {
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}
	ms := storage.NewMemStorage()
	wg := new(sync.WaitGroup)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(4)

	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	pollTicker := time.NewTicker(config.PollInterval)
	c := sync.NewCond(&sync.Mutex{})
	go pollScheduller(ctx, c, pollTicker, wg)
	go pollRuntimeMetrics(ctx, c, ms, wg)
	go pollPSUtilMetrics(ctx, c, ms, wg)
	go reportMetrics(ctx, &config, ms, wg)
	sig := <-syscallCancelChan
	log.Printf("Caught syscall: %v", sig)
	cancel()
	c.Broadcast()
	wg.Wait()
}
