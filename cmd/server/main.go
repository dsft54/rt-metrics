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
	"github.com/gin-gonic/gin"

	"github.com/dsft54/rt-metrics/cmd/server/handlers"
	"github.com/dsft54/rt-metrics/cmd/server/settings"
	"github.com/dsft54/rt-metrics/cmd/server/storage"
)

var (
	config    settings.Config
	memstore  storage.MemoryStorage
	filestore storage.FileStorage
	dbstore   storage.DBStorage
)

func init() {
	memstore = storage.MemoryStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}

	flag.StringVar(&config.Address, "a", "localhost:8080", "Server address")
	flag.BoolVar(&config.Restore, "r", true, "Restore metrics from file on start")
	flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "Path to file storage")
	flag.DurationVar(&config.StoreInterval, "i", 300*time.Second, "Update file storage interval")
	flag.StringVar(&config.HashKey, "k", "", "SHA256 signing key")
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:example@localhost:5432/postgres", "Postgress connection uri")
}

func setupGinHandlers() *gin.Engine {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		handlers.Decompression(),
		handlers.Compression(),
		gin.Logger(),
	)

	router.GET("/", handlers.RootHandler(&memstore))
	router.GET("/ping", handlers.PingDB(&dbstore))
	router.GET("/value/:type/:name", handlers.AddressedRequest(&memstore))
	router.POST("/update/", handlers.HandleUpdateJSON(&filestore, &memstore, config.HashKey))
	router.POST("/value/", handlers.HandleRequestJSON(&memstore, config.HashKey))
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)
	router.POST("/update/:type/:name/:value", handlers.StringUpdatesHandler(&filestore, &memstore))

	return router
}

func main() {
	// Init syscall chanel, ctx, parse flags and os vars
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		log.Println(err)
	}
	// Init file and db storages
	filestore.InitFileStorage(config)
	err = dbstore.SetupDBStorage(config.DatabaseDSN, ctx)
	if err != nil {
		log.Println("DB: ", err)
	}
	defer dbstore.Connection.Close()
	// Handle file interaction if neccesary
	if config.Restore {
		err := memstore.ReadOldMetrics(filestore.FilePath)
		if err != nil {
			log.Println("Wanted to restore old metrics from file on server start but failed; ", err)
		}
	}
	if filestore.StoreData && !filestore.Synchronize {
		go filestore.IntervalUpdate(ctx, config.StoreInterval, &memstore)
	}

	router := setupGinHandlers()
	server := &http.Server{
		Addr:    config.Address,
		Handler: router,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Println("Listen: ", err)
		}
	}()

	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-syscallCancelChan
	log.Println("Caught syscall: ", sig)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
	cancel()
	err = filestore.SaveDataToFile(filestore.StoreData, &memstore)
	if err != nil {
		log.Println("Failed to save data on server exit;", err)
	}
}
