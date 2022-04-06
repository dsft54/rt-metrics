package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dsft54/rt-metrics/cmd/server/storage"
	"github.com/gin-gonic/gin"
)

func RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Gauge metrics: %+v\n, Counter metrics: %+v\n", storage.Store.GaugeMetrics, storage.Store.CounterMetrics)
}

func AddressedRequest(c *gin.Context) {
	mType := c.Param("type")
	mName := c.Param("name")
	switch mType {
	case "gauge":
		if _, found := storage.Store.GaugeMetrics[mName]; found {
			c.String(http.StatusOK, "%v", storage.Store.GaugeMetrics[mName])
			return
		}
		c.Status(http.StatusNotFound)
		return
	case "counter":
		if _, found := storage.Store.CounterMetrics[mName]; found {
			c.String(http.StatusOK, "%v", storage.Store.CounterMetrics[mName])
			return
		}
		c.Status(http.StatusNotFound)
		return
	default:
		c.Status(http.StatusNotFound)
		return
	}
}

func HandleRequestJSON(c *gin.Context) {
	rawData, err := c.GetRawData()
	if err != nil {
		fmt.Println(err)
	}
	metricsRequest := &storage.Metrics{}
	err = json.Unmarshal(rawData, metricsRequest)
	if err != nil {
		fmt.Println(err)
	}
	switch metricsRequest.MType {
	case "gauge":
		v := storage.Store.GaugeMetrics[metricsRequest.ID]
		metricsRequest.Value = &v
	case "counter":
		v := storage.Store.CounterMetrics[metricsRequest.ID]
		metricsRequest.Delta = &v
	}
	c.JSON(http.StatusOK, metricsRequest)
}

func HandleUpdateJSON(c *gin.Context) {
	rawData, err := c.GetRawData()
	if err != nil {
		fmt.Println(err)
	}
	metricsRequest := &storage.Metrics{}
	err = json.Unmarshal(rawData, metricsRequest)
	if err != nil {
		fmt.Println(err)
	}
	switch metricsRequest.MType {
	case "gauge":
		storage.Store.GaugeMetrics[metricsRequest.ID] = *metricsRequest.Value
	case "counter":
		storage.Store.CounterMetrics[metricsRequest.ID] = *metricsRequest.Delta
	}
	c.Status(http.StatusOK)
}

func WithoutID(c *gin.Context) {
	c.Status(http.StatusNotFound)
}
