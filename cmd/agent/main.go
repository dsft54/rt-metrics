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
	
	"github.com/dsft54/rt-metrics/config/agent/settings"
	"github.com/dsft54/rt-metrics/internal/agent/scheduller"
	"github.com/dsft54/rt-metrics/internal/agent/storage"
)

// sendData собирает json в массив байт, и отправляет его при помощи resty.Client на
// url в теле POST запроса.
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

// reportMetrics ожидает либо выхода по контексту, либо бродкаста на переменную состояния. Во втором
// случае отправляет метрики на сервер либо штучно, либо списком.
func reportMetrics(ctx context.Context, sch *scheduller.Scheduller, cfg *settings.Config, s *storage.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	client := resty.New()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			sch.Rc.L.Lock()
			sch.Rc.Wait()
			sch.Rc.L.Unlock()
			if !sch.Update {
				return
			}
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
		}
	}
}

// pollRuntimeMetrics ожидает либо выхода по контексту, либо бродкаста на переменную состояния. Во втором случае
// собирает в хранилище в памяти Runtime Metrics.
func pollRuntimeMetrics(ctx context.Context, c *sync.Cond, s *storage.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.CollectRuntimeMetrics()
			log.Println("All runtime memory stats collected")
			c.L.Lock()
			c.Wait()
			c.L.Unlock()
		}
	}
}

// pollPSUtilMetrics ожидает либо выхода по контексту, либо бродкаста на переменную состояния. Во втором случае
// собирает в хранилище в памяти метрики процессора и памяти.
func pollPSUtilMetrics(ctx context.Context, c *sync.Cond, s *storage.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.CollectPSUtilMetrics()
			log.Println("All psutil memory stats collected")
			c.L.Lock()
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
	sch := scheduller.NewScheduller(&config)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(4)

	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go sch.Start(ctx, wg)
	go pollRuntimeMetrics(ctx, sch.Pc, ms, wg)
	go pollPSUtilMetrics(ctx, sch.Pc, ms, wg)
	go reportMetrics(ctx, sch, &config, ms, wg)
	sig := <-syscallCancelChan
	log.Printf("Caught syscall: %v", sig)
	cancel()
	sch.ExitRelease()
	wg.Wait()
}
