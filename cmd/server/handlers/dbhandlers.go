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

func DBStringUpdatesHandler(db *storage.DBStorage, fs *storage.FileStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		mType := c.Param("type")
		mName := c.Param("name")
		mValue := c.Param("value")
		code, err := db.DBUpdateValueFromParams(mType, mName, mValue)
		if err != nil {
			log.Println(err, "Type:", mType, "Name:", mName, "Value:", mValue, "Code:", code)
		}
		if code == 200 {
			err = db.DBSaveToFile(fs)
			if err != nil {
				log.Println("Synchronized data saving was failed", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		c.Status(code)
	}
}

func DBAddressedRequest(db *storage.DBStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		rType := c.Param("type")
		rID := c.Param("name")
		metricsRequest := storage.Metrics{
			ID:    rID,
			MType: rType,
		}
		metricsResponse, err := db.DBReadSpecific(&metricsRequest)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		if rType == "counter" {
			c.String(http.StatusOK, "%v", *metricsResponse.Delta)
			return
		}
		if rType == "gauge" {
			c.String(http.StatusOK, "%v", *metricsResponse.Value)
			return
		}
		c.Status(http.StatusNotFound)
		return
	}
}

func DBHandleRequestJSON(db *storage.DBStorage, key string) gin.HandlerFunc {
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
		metricsResponse, err := db.DBReadSpecific(metricsRequest)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		switch metricsResponse.MType {
		case "gauge":
			h := hmac.New(sha256.New, []byte(key))
			h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metricsResponse.ID, *metricsResponse.Value)))
			metricsResponse.Hash = hex.EncodeToString(h.Sum(nil))
		case "counter":
			h := hmac.New(sha256.New, []byte(key))
			h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricsResponse.ID, *metricsResponse.Delta)))
			metricsResponse.Hash = hex.EncodeToString(h.Sum(nil))
		}
		c.JSON(http.StatusOK, metricsResponse)
	}
}

func DBHandleUpdateJSON(db *storage.DBStorage, fs *storage.FileStorage, key string) gin.HandlerFunc {
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
			err = db.DBInsertGauge(metricsRequest)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		case "counter":
			if key != "" {
				h := hmac.New(sha256.New, []byte(key))
				h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricsRequest.ID, *metricsRequest.Delta)))
				if metricsRequest.Hash != hex.EncodeToString(h.Sum(nil)) {
					c.Status(http.StatusBadRequest)
					return
				}
			}
			err = db.DBInsertCounter(metricsRequest)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		err = db.DBSaveToFile(fs)
		if err != nil {
			log.Println("Synchronized data saving was failed", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func PingDB(db *storage.DBStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db.Connection == nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		err := db.Ping()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func DBRootHandler(db *storage.DBStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics, err := db.DBReadAll()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		data, err := json.Marshal(metrics)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		return
	}
}

func DBBatchUpdate(db *storage.DBStorage, fs *storage.FileStorage, key string) gin.HandlerFunc {
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
		metricsBatchClean := []storage.Metrics{}
		if key != "" {
			for _, metric := range metricsBatch {
				switch metric.MType {
				case "gauge":
					h := hmac.New(sha256.New, []byte(key))
					h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value)))
					if metric.Hash != hex.EncodeToString(h.Sum(nil)) {
						continue
					}
					metricsBatchClean = append(metricsBatchClean, metric)
				case "counter":
					h := hmac.New(sha256.New, []byte(key))
					h.Write([]byte(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta)))
					if metric.Hash != hex.EncodeToString(h.Sum(nil)) {
						continue
					}
					metricsBatchClean = append(metricsBatchClean, metric)
				}
			}
			metricsBatch = metricsBatchClean
		}
		err = db.DBBatchQuery(metricsBatch)
		if err != nil {
			log.Println("Error while update metrics from batch", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		err = db.DBSaveToFile(fs)
		if err != nil {
			log.Println("Synchronized data saving was failed", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}
