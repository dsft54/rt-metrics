package main

import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/dsft54/rt-metrics/config/server/settings"
	"github.com/dsft54/rt-metrics/internal/cryptokey"
	"github.com/dsft54/rt-metrics/internal/server/grpcs"
	"github.com/dsft54/rt-metrics/internal/server/handlers"
	"github.com/dsft54/rt-metrics/internal/server/storage"
	pb "github.com/dsft54/rt-metrics/proto"
)

var config settings.Config

// initStorages в зависимости от успеха подключения к бд выбирает активный storage, куда будут сохраняться метрики,
// а также создает файловое хранище на основе настроек сервера.
func initStorages(ctx context.Context, config settings.Config) (storage.IStorage, *storage.FileStorage) {
	// Init file and db storages
	filestore := storage.NewFileStorage(config)
	dbstore := &storage.DBStorage{}
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

// setupGinRouter создает *gin.Engine определяя работу маршрутизатора и используемое middleware.
func setupGinRouter(st storage.IStorage, fs *storage.FileStorage, keyPath, allowedNetwork string) (*gin.Engine, error) {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		handlers.Decompression(),
		handlers.Compression(gzip.BestSpeed),
		gin.Logger(),
	)
	if allowedNetwork != "" {
		router.Use(handlers.NetFilter(allowedNetwork))
	}
	if keyPath != "" {
		private, err := cryptokey.ParsePrivateKey(keyPath)
		if err != nil {
			return nil, err
		}
		pub, err := cryptokey.ParsePublicKey(keyPath + ".pub")
		if err != nil {
			return nil, err
		}
		chunkSize := pub.Size()
		router.Use(
			handlers.Decryption(private, chunkSize),
		)
	}
	router.GET("/", handlers.RequestAllMetrics(st))
	router.GET("/ping", handlers.PingDatabase(st))
	router.GET("/value/:type/:name", handlers.AddressedRequest(st))
	router.POST("/value/", handlers.RequestMetricJSON(st, config.HashKey))
	router.POST("/update/", handlers.UpdateMetricJSON(st, fs, config.HashKey))
	router.POST("/updates/", handlers.BatchUpdateJSON(st, fs, config.HashKey))
	router.POST("/update/:type/:name/:value", handlers.ParametersUpdate(st, fs))
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)
	return router, nil
}

// init определяет используемые флаги командной строки для настройки запуска сервера.
func init() {
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://postgres:example@localhost:5432", "Postgress connection uri")
	flag.StringVar(&config.StoreFile, "f", "/tmp/devops-metrics-db.json", "Path to file storage")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "Path to public rsa key")
	flag.StringVar(&config.TrustedSubnet, "t", "", "Trusted source of logs")
	flag.StringVar(&config.Address, "a", "localhost:8080", "Server address")
	flag.StringVar(&config.Config, "c", "", "Path to json config file")
	flag.StringVar(&config.HashKey, "k", "", "SHA256 signing key")
	flag.BoolVar(&config.Restore, "r", true, "Restore metrics from file on start")
	flag.BoolVar(&config.Grpc, "g", true, "Start up grpc server")
	flag.DurationVar(&config.StoreInterval, "i", 300*time.Second, "Update file storage interval")
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
	err = config.ParseFromFile()
	if err != nil {
		log.Println(err)
	}
	st, fs := initStorages(ctx, config)
	log.Println("Running config - ", config)

	// Handle file interaction if neccesary
	if config.Restore {
		err = st.UploadFromFile(fs.FilePath)
		if err != nil {
			log.Println("Wanted to restore old metrics from file on server start but failed; ", err)
		}
	}
	if fs.StoreData && !fs.Synchronize {
		go fs.IntervalUpdate(ctx, config.StoreInterval, st)
	}

	// Start gin engine
	router, err := setupGinRouter(st, fs, config.CryptoKey, config.TrustedSubnet)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:    config.Address,
		Handler: router,
	}
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Println("Listen: ", err)
		}
	}()

	// Start grpc server
	if config.Grpc {
		listen, err := net.Listen("tcp", ":3200")
		if err != nil {
			log.Fatal(err)
		}
		s := grpc.NewServer()
		pb.RegisterMetricServer(s, &grpcs.MetricsServer{Storage: st})

		fmt.Println("Сервер gRPC начал работу")
		if err := s.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}

	// Wait and handle syscall exits
	signal.Notify(syscallCancelChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-syscallCancelChan
	log.Println("Caught syscall: ", sig)
	if err = server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

	// Collect memory profile
	fmem, err := os.Create(`profiles/new_server_mem.profile`)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC()
	if err = pprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}

	// Store data in file on exit if condition
	if fs.StoreData {
		err = fs.SaveStorageToFile(st)
		if err != nil {
			log.Println("Failed to save data on server exit;", err)
		}
		log.Println("Saved db to file on exit")
	}
}
