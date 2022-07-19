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

var config settings.Config

func initStorages(ctx context.Context, config settings.Config) (storage.Storage, *storage.FileStorage) {
	// Init file and db storages
	filestore := storage.NewFileStorage(config)
	dbstore := new(storage.DBStorage)
	err := dbstore.DBConnectStorage(ctx, config.DatabaseDSN)
	if err != nil {
		log.Println("DB error : ", err)
	}
	if dbstore.Connection != nil {
		log.Println("DB connection: Success")
		return dbstore, filestore
	}
	memstore := storage.MemoryStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
	return &memstore, filestore
}

func setupGinRouter(st storage.Storage, fs *storage.FileStorage) *gin.Engine {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		handlers.Decompression(),
		handlers.Compression(),
		gin.Logger(),
	)
	router.GET("/", handlers.RequestAllMetrics(st))
	router.GET("/ping", handlers.PingDatabase(st))
	router.GET("/value/:type/:name", handlers.AddressedRequest(st))
	router.POST("/value/", handlers.RequestMetricJSON(st, config.HashKey))
	router.POST("/update/", handlers.UpdateMetricJSON(st, fs, config.HashKey))
	router.POST("/updates/", handlers.BatchUpdateJSON(st, fs, config.HashKey))
	router.POST("/update/:type/:name/:value", handlers.ParametersUpdate(st, fs))
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)
	return router
}

func init() {
	flag.StringVar(&config.Address, "a", "localhost:8080", "Server address")
	flag.BoolVar(&config.Restore, "r", true, "Restore metrics from file on start")
	flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "Path to file storage")
	flag.DurationVar(&config.StoreInterval, "i", 300*time.Second, "Update file storage interval")
	flag.StringVar(&config.HashKey, "k", "", "SHA256 signing key")
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:example@localhost:5432", "Postgress connection uri")
}

func main() {
	// Init syscall channel, ctx, stores, parse flags and os vars
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		log.Println(err)
	}
	st, fs := initStorages(ctx, config)
	log.Println("Running config - ", config)

	// Handle file interaction if neccesary
	if config.Restore {
		err := st.UploadFromFile(fs.FilePath)
		if err != nil {
			log.Println("Wanted to restore old metrics from file on server start but failed; ", err)
		}
	}
	if fs.StoreData && !fs.Synchronize {
		go fs.IntervalUpdate(ctx, config.StoreInterval, st)
	}

	// Start gin engine
	router := setupGinRouter(st, fs)
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

	// Wait and handle syscall exits
	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-syscallCancelChan
	log.Println("Caught syscall: ", sig)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

	// Store data in file on exit if condition
	if fs.StoreData {
		err = fs.SaveStorageToFile(st)
		if err != nil {
			log.Println("Failed to save data on server exit;", err)
		}
		log.Println("Saved db to file on exit")
	}
}
