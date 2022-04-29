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
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:example@localhost:5432", "Postgress connection uri")
}

func setupGinHandlers() *gin.Engine {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		handlers.Decompression(),
		handlers.Compression(),
		gin.Logger(),
	)

	if dbstore.Connection != nil {
		router.GET("/", handlers.DBRootHandler(&dbstore))
		router.GET("/value/:type/:name", handlers.DBAddressedRequest(&dbstore))
		router.POST("/update/", handlers.DBHandleUpdateJSON(&dbstore, &filestore, config.HashKey))
		router.POST("/value/", handlers.DBHandleRequestJSON(&dbstore, config.HashKey))
		router.POST("/updates/", handlers.DBBatchUpdate(&dbstore, &filestore, config.HashKey))
		router.POST("/update/:type/:name/:value", handlers.DBStringUpdatesHandler(&dbstore, &filestore))
	} else {
		router.GET("/", handlers.RootHandler(&memstore))
		router.GET("/value/:type/:name", handlers.AddressedRequest(&memstore))
		router.POST("/update/", handlers.HandleUpdateJSON(&memstore, &filestore, config.HashKey))
		router.POST("/value/", handlers.HandleRequestJSON(&memstore, config.HashKey))
		router.POST("/updates/", handlers.BatchUpdate(&memstore, &filestore, config.HashKey))
		router.POST("/update/:type/:name/:value", handlers.StringUpdatesHandler(&memstore, &filestore))
	}
	router.GET("/ping", handlers.PingDB(&dbstore))
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)

	return router
}

func main() {
	// Init syscall chanel, ctx, parse flags and os vars
	syscallCancelChan := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		log.Println(err)
	}
	log.Println("Running config - ", config)

	// Init file and db storages
	filestore.InitFileStorage(config)
	err = dbstore.DBConnectStorage(ctx, config.DatabaseDSN, "rt_metrics")
	if err != nil {
		log.Println("DB error : ", err)
	}
	if dbstore.Connection != nil {
		defer dbstore.Connection.Close()
		log.Println("DB connection: Success")
		err := dbstore.DBCreateTable()
		if err != nil {
			log.Println("DB failed to create table, ", dbstore.TableName, err)
		}
	}

	// Handle file interaction if neccesary
	if config.Restore {
		if dbstore.Connection != nil {
			err := dbstore.ReadOldMetrics(filestore.FilePath)
			if err != nil {
				log.Println("(DBStorage) Wanted to restore old metrics from file on server start but failed; ", err)
			}
		} else {
			err := memstore.ReadOldMetrics(filestore.FilePath)
			if err != nil {
				log.Println("(Memstorage) Wanted to restore old metrics from file on server start but failed; ", err)
			}
		}
	}
	if filestore.StoreData && !filestore.Synchronize && config.DatabaseDSN == "" {
		go filestore.IntervalUpdateMem(ctx, config.StoreInterval, &memstore)
	}
	if filestore.StoreData && !filestore.Synchronize && config.DatabaseDSN != "" {
		go filestore.IntervalUpdateDB(ctx, config.StoreInterval, &dbstore)
	}

	// Start gin engine
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

	// Wait and handle syscall exits
	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-syscallCancelChan
	log.Println("Caught syscall: ", sig)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

	// Store data in file on exit if condition
	if dbstore.Connection != nil {
		err = filestore.SaveDBDataToFile(filestore.StoreData, &dbstore)
		if err != nil {
			log.Println("Failed to save data on server exit (DBStorage);", err)
		}
		log.Println("SAVED DB on exit")
	} else {
		err = filestore.SaveMemDataToFile(filestore.StoreData, &memstore)
		if err != nil {
			log.Println("Failed to save data on server exit (MEMStorage);", err)
		}
		log.Println("SAVED MEM on exit")
	}
}
