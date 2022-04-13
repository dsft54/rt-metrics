package main

import (
	"flag"
	"fmt"

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

	err := env.Parse(&settings.Cfg)
	if err != nil {
		fmt.Println(err)
	}

	flag.StringVar(&settings.Cfg.Address,"a", settings.Cfg.Address, "Server address")
	flag.BoolVar(&settings.Cfg.Restore, "r", settings.Cfg.Restore, "Restore metrics from file on start")
	flag.StringVar(&settings.Cfg.StoreFile, "f", settings.Cfg.StoreFile, "Path to file storage")
	flag.DurationVar(&settings.Cfg.StoreInterval,"i", settings.Cfg.StoreInterval, "Update file storage interval")
	
	err = storage.FileStore.InitFileStorage(settings.Cfg)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	if !storage.FileStore.Synchronize {
		go storage.FileStore.IntervalUpdate(settings.Cfg.StoreInterval)
	}
	flag.Parse()

	router := gin.New()
	router.Use(gin.Recovery(), handlers.Decompression(), handlers.Compression())
	// router := gin.Default()

	// router.POST("/update/:type/:name/:value", handlers.UpdatesHandler)
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)
	router.POST("/update", handlers.HandleUpdateJSON)

	router.POST("/value", handlers.HandleRequestJSON)
	router.GET("/", handlers.RootHandler)
	router.GET("/value/:type/:name", handlers.AddressedRequest)

	router.Run(settings.Cfg.Address)
}
