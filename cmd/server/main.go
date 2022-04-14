package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/dsft54/rt-metrics/cmd/server/handlers"
	"github.com/dsft54/rt-metrics/cmd/server/settings"
	"github.com/dsft54/rt-metrics/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

func init() {
	storage.Store = storage.MetricStorages{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}

	flag.StringVar(&settings.Cfg.Address, "a", "localhost:8080", "Server address")
	flag.BoolVar(&settings.Cfg.Restore, "r", true, "Restore metrics from file on start")
	flag.StringVar(&settings.Cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "Path to file storage")
	flag.DurationVar(&settings.Cfg.StoreInterval, "i", 300*time.Second, "Update file storage interval")
}

func setupGinHandlers() *gin.Engine {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		handlers.Decompression(),
		handlers.Compression(),
		gin.Logger(),
	)

	router.GET("/", handlers.RootHandler)
	router.GET("/value/:type/:name", handlers.AddressedRequest)
	router.POST("/update/", handlers.HandleUpdateJSON)
	router.POST("/value/", handlers.HandleRequestJSON)
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)
	router.POST("/update/:type/:name/:value", handlers.StringUpdatesHandler)

	return router
}

func main() {
	flag.Parse()
	err := env.Parse(&settings.Cfg)
	if err != nil {
		log.Println(err)
	}
	err = storage.FileStore.InitFileStorage(settings.Cfg, &storage.Store)
	if err != nil {
		log.Println(err, " - file specified for restoring metrics was not found")
	}
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	if storage.FileStore.StoreData && !storage.FileStore.Synchronize {
		go storage.FileStore.IntervalUpdate(settings.Cfg.StoreInterval, &storage.Store, ctx)
	}

	router := setupGinHandlers()
	server := &http.Server{
		Addr:    settings.Cfg.Address,
		Handler: router,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("\nListen: %s\n", err)
		}
	}()

	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-syscallCancelChan
	defer cancel()
	log.Printf("\nCaught syscall: %v\n", sig)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

	if storage.FileStore.StoreData {
		err := storage.FileStore.OpenToWrite()
		if err != nil {
			log.Println("Failed to save data on exit")
		}
		storage.Store.WriteMetricsToFile(storage.FileStore.File)
		storage.FileStore.File.Close()
		log.Println("Data was saved")
	}
}
