package main

import (
	"github.com/dsft54/rt-metrics/cmd/server/handlers"
	"github.com/dsft54/rt-metrics/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

func init() {
	storage.Store = storage.MetricStorages{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func main() {
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

	router.Run(":8080")
}
