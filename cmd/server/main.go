package main

import (
	"github.com/caarlos0/env"
	"github.com/dsft54/rt-metrics/cmd/server/handlers"
	"github.com/dsft54/rt-metrics/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func init() {
	storage.Store = storage.MetricStorages{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func main() {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	// router := gin.New()
	// router.Use(gin.Recovery())
	router := gin.Default()

	// router.POST("/update/:type/:name/:value", handlers.UpdatesHandler)
	router.POST("/update/gauge/", handlers.WithoutID)
	router.POST("/update/counter/", handlers.WithoutID)
	router.POST("/update", handlers.HandleUpdateJSON)

	router.POST("/value", handlers.HandleRequestJSON)
	router.GET("/", handlers.RootHandler)
	router.GET("/value/:type/:name", handlers.AddressedRequest)

	router.Run(cfg.Address)
}
