package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dsft54/rt-metrics/cmd/server/storage"
)

func StringUpdatesHandler(st *storage.MemoryStorage, fs *storage.FileStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		mType := c.Param("type")
		mName := c.Param("name")
		mValue := c.Param("value")
		code, err := st.UpdateMetricsFromString(mType, mName, mValue)
		if err != nil {
			log.Println(err, "Type:", mType, "Name:", mName, "Value:", mValue, "Code", code)
		}
		if code == 200 {
			err := fs.SaveMemDataToFile(fs.Synchronize, st)
			if err != nil {
				log.Println("Synchronized data saving was failed", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		c.Status(code)
	}
}

func AddressedRequest(st *storage.MemoryStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		rType := c.Param("type")
		rID := c.Param("name")
		switch rType {
		case "gauge":
			if _, found := st.GaugeMetrics[rID]; found {
				c.String(http.StatusOK, "%v", st.GaugeMetrics[rID])
				return
			}
			c.Status(http.StatusNotFound)
			return
		case "counter":
			if _, found := st.CounterMetrics[rID]; found {
				c.String(http.StatusOK, "%v", st.CounterMetrics[rID])
				return
			}
			c.Status(http.StatusNotFound)
			return
		default:
			c.Status(http.StatusNotFound)
			return
		}
	}
}

func HandleRequestJSON(st *storage.MemoryStorage, key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawData, err := c.GetRawData()
		if err != nil {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		metricsRequest := &storage.Metrics{}
		err = json.Unmarshal(rawData, metricsRequest)
		if err != nil {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		switch metricsRequest.MType {
		case "gauge":
			v := st.GaugeMetrics[metricsRequest.ID]
			metricsRequest.Value = &v
			h := hmac.New(sha256.New, []byte(key))
			h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metricsRequest.ID, v)))
			metricsRequest.Hash = hex.EncodeToString(h.Sum(nil))
		case "counter":
			d := st.CounterMetrics[metricsRequest.ID]
			metricsRequest.Delta = &d
			h := hmac.New(sha256.New, []byte(key))
			h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricsRequest.ID, d)))
			metricsRequest.Hash = hex.EncodeToString(h.Sum(nil))
		}
		c.JSON(http.StatusOK, metricsRequest)
	}
}

func HandleUpdateJSON(st *storage.MemoryStorage, fs *storage.FileStorage, key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawData, err := c.GetRawData()
		if err != nil {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		metricsRequest := &storage.Metrics{}
		err = json.Unmarshal(rawData, metricsRequest)
		if err != nil {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		switch metricsRequest.MType {
		case "gauge":
			if key != "" {
				h := hmac.New(sha256.New, []byte(key))
				h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metricsRequest.ID, *metricsRequest.Value)))
				if metricsRequest.Hash != hex.EncodeToString(h.Sum(nil)) {
					c.Status(http.StatusBadRequest)
					return
				}
			}
			st.GaugeMetrics[metricsRequest.ID] = *metricsRequest.Value
		case "counter":
			if key != "" {
				h := hmac.New(sha256.New, []byte(key))
				h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricsRequest.ID, *metricsRequest.Delta)))
				if metricsRequest.Hash != hex.EncodeToString(h.Sum(nil)) {
					c.Status(http.StatusBadRequest)
					return
				}
			}
			st.CounterMetrics[metricsRequest.ID] += *metricsRequest.Delta
		}
		err = fs.SaveMemDataToFile(fs.Synchronize, st)
		if err != nil {
			log.Println("Synchronized data saving was failed", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func BatchUpdate(st *storage.MemoryStorage, fs *storage.FileStorage, key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawData, err := c.GetRawData()
		if err != nil {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		metricsBatch := []storage.Metrics{}
		err = json.Unmarshal(rawData, &metricsBatch)
		if err != nil {
			log.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		for _, metric := range metricsBatch {
			switch metric.MType {
			case "gauge":
				if key != "" {
					h := hmac.New(sha256.New, []byte(key))
					h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value)))
					if metric.Hash != hex.EncodeToString(h.Sum(nil)) {
						continue
					}
				}
				st.GaugeMetrics[metric.ID] = *metric.Value
			case "counter":
				if key != "" {
					h := hmac.New(sha256.New, []byte(key))
					h.Write([]byte(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta)))
					if metric.Hash != hex.EncodeToString(h.Sum(nil)) {
						continue
					}
				}
				st.CounterMetrics[metric.ID] += *metric.Delta
			}
		}
		err = fs.SaveMemDataToFile(fs.Synchronize, st)
		if err != nil {
			log.Println("Synchronized data saving was failed", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func WithoutID(c *gin.Context) {
	c.Status(http.StatusNotFound)
}

func RootHandler(st *storage.MemoryStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf("Gauge metrics: %+v\n, Counter metrics: %+v\n", st.GaugeMetrics, st.CounterMetrics)))
	}
}
