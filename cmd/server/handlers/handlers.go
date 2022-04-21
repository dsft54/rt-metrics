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

func StringUpdatesHandler(fs *storage.FileStorage, st *storage.MemoryStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		mType := c.Param("type")
		mName := c.Param("name")
		mValue := c.Param("value")
		code, err := st.UpdateMetricsFromString(mType, mName, mValue)
		if err != nil {
			log.Println(err, "Type:", mType, "Name:", mName, "Value:", mValue)
		}
		switch code {
		case 200:
			err := fs.SaveDataToFile(fs.Synchronize, st)
			if err != nil {
				log.Println("Synchronized data saving was failed", err)
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusOK)
		case 400:
			c.Status(http.StatusBadRequest)
		case 501:
			c.Status(http.StatusNotImplemented)
		}
	}
}

func AddressedRequest(st *storage.MemoryStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		mType := c.Param("type")
		mName := c.Param("name")
		switch mType {
		case "gauge":
			if _, found := st.GaugeMetrics[mName]; found {
				c.String(http.StatusOK, "%v", st.GaugeMetrics[mName])
				return
			}
			c.Status(http.StatusNotFound)
			return
		case "counter":
			if _, found := st.CounterMetrics[mName]; found {
				c.String(http.StatusOK, "%v", st.CounterMetrics[mName])
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

func HandleUpdateJSON(fs *storage.FileStorage, st *storage.MemoryStorage, key string) gin.HandlerFunc {
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
		err = fs.SaveDataToFile(fs.Synchronize, st)
		if err != nil {
			log.Println("Synchronized data saving was failed", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func PingDB(dbs *storage.DBStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if dbs.Connection == nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		err := dbs.Ping()
		if err != nil {
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
