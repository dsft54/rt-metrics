package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	gaugeMetrics   map[string]float64 = make(map[string]float64)
	counterMetrics map[string]int     = make(map[string]int)
)

func UpdateMetrics(metricType, metricName, metricValue string) (int, error) {
	switch metricType {
	case "gauge":
		floatFromString, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return 400, err
		}
		gaugeMetrics[metricName] = floatFromString
		return 200, nil
	case "counter":
		intFromString, err := strconv.Atoi(metricValue)
		if err != nil {
			return 400, err
		}
		counterMetrics[metricName] += intFromString
		return 200, nil
	default:
		return 501, errors.New("wrong metric type - " + metricType)
	}
}

func UpdatesHandler(c *gin.Context) {
	mType := c.Param("type")
	mName := c.Param("name")
	mValue := c.Param("value")
	code, err := UpdateMetrics(mType, mName, mValue)
	if err != nil {
		fmt.Println(err, "Type:", mType, "Name:", mName, "Value:", mValue)
	}
	switch code {
	case 200:
		c.Status(http.StatusOK)
	case 400:
		c.Status(http.StatusBadRequest)
	case 501:
		c.Status(http.StatusNotImplemented)
	}
}

func WithoutID(c *gin.Context) {
	c.Status(http.StatusNotFound)
}

func AddressedRequest(c *gin.Context) {
	mType := c.Param("type")
	mName := c.Param("name")
	switch mType {
	case "gauge":
		if _, found := gaugeMetrics[mName]; found {
			c.String(http.StatusOK, "%v", gaugeMetrics[mName])
			return
		}
		c.Status(http.StatusNotFound)
		return
	case "counter":
		if _, found := counterMetrics[mName]; found {
			c.String(http.StatusOK, "%v", counterMetrics[mName])
			return
		}
		c.Status(http.StatusNotFound)
		return
	default:
		c.Status(http.StatusNotFound)
		return
	}
}

func RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Gauge metrics: %+v\n, Counter metrics: %+v\n", gaugeMetrics, counterMetrics)
}

func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/update/:type/:name/:value", UpdatesHandler)
	router.POST("/update/gauge/", WithoutID)
	router.POST("/update/counter/", WithoutID)

	router.GET("/", RootHandler)
	router.GET("/value/:type/:name", AddressedRequest)

	router.Run(":8080")
}
