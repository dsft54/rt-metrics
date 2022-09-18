package main

import (
	"context"
	"crypto/rsa"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dsft54/rt-metrics/config/agent/settings"
	"github.com/dsft54/rt-metrics/internal/agent/grpcc"
	"github.com/dsft54/rt-metrics/internal/agent/scheduller"
	"github.com/dsft54/rt-metrics/internal/agent/storage"
	"github.com/dsft54/rt-metrics/internal/cryptokey"
	pb "github.com/dsft54/rt-metrics/proto"
)

// sendData собирает json в массив байт, и отправляет его при помощи resty.Client на
// url в теле POST запроса.
func sendData(url string, keyPath string, m interface{}, client *resty.Client) error {
	rawData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if keyPath != "" {
		var pub *rsa.PublicKey
		pub, err = cryptokey.ParsePublicKey(keyPath)
		if err != nil {
			return err
		}
		rawData, err = cryptokey.EncryptMessage(rawData, pub)
		if err != nil {
			return err
		}
	}
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Real-IP", "127.0.0.1").
		SetBody(rawData).
		Post(url)
	return err
}

// reportMetrics ожидает либо выхода по контексту, либо бродкаста на переменную состояния. Во втором
// случае отправляет метрики на сервер либо штучно, либо списком.
func reportMetrics(ctx context.Context, sch *scheduller.Scheduller, cfg *settings.Config, s *storage.MemStorage, c pb.MetricClient, wg *sync.WaitGroup) {
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
						err := sendData("http://"+cfg.Address+"/update", cfg.CryptoKey, &value, client)
						if err != nil {
							log.Println(err)
							continue
						}
						if cfg.Grpc {
							err := grpcc.SendMetric(ctx, c, value)
							if err != nil {
								log.Println(err)
								continue
							}
						}
					}
				}
			} else {
				err := sendData("http://"+cfg.Address+"/updates", cfg.CryptoKey, &metricsSlice, client)
				if err != nil {
					log.Println(err)
				}
				if cfg.Grpc {
					err := grpcc.SendMetrics(ctx, c, metricsSlice)
					if err != nil {
						log.Println(err)
					}
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
	flag.BoolVar(&config.Grpc, "g", true, "Send metrics via grpc")
	flag.StringVar(&config.HashKey, "k", "", "SHA256 signing key")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Path to public rsa key")
	flag.StringVar(&config.Config, "c", "", "Path to json config file")
}

var config       settings.Config


func main() {
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}
	err = config.ParseFromFile()
	if err != nil {
		log.Println(err)
	}
	ms := storage.NewMemStorage()
	wg := new(sync.WaitGroup)
	sch := scheduller.NewScheduller(&config)
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(4)

	var c pb.MetricClient
	if config.Grpc {
		conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		c = pb.NewMetricClient(conn)
		log.Println("Connection to grpc server established")
	}

	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go sch.Start(ctx, wg)
	go pollRuntimeMetrics(ctx, sch.Pc, ms, wg)
	go pollPSUtilMetrics(ctx, sch.Pc, ms, wg)
	go reportMetrics(ctx, sch, &config, ms, c, wg)
	sig := <-syscallCancelChan
	log.Printf("Caught syscall: %v", sig)
	cancel()
	sch.ExitRelease()
	wg.Wait()
}
