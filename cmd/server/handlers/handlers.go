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

func ParametersUpdate(st storage.Storage, fs *storage.FileStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		mType := c.Param("type")
		mName := c.Param("name")
		mValue := c.Param("value")
		code, err := st.ParamsUpdate(mType, mName, mValue)
		if err != nil {
			log.Println(err, "Type:", mType, "Name:", mName, "Value:", mValue, "Code", code)
		}
		if code == 200 {
			if fs.Synchronize {
				err := fs.SaveStorageToFile(st)
				if err != nil {
					log.Println("Synchronized data saving was failed", err)
					c.Status(http.StatusInternalServerError)
					return
				}
			}
		}
		c.Status(code)
	}
}

func AddressedRequest(st storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		rType := c.Param("type")
		rID := c.Param("name")
		metricsRequest := storage.Metrics{
			ID:    rID,
			MType: rType,
		}
		metricsResponse, err := st.ReadMetric(&metricsRequest)
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
	}
}

func RequestMetricJSON(st storage.Storage, key string) gin.HandlerFunc {
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
		log.Println("Request JSON ----- ", metricsRequest)
		metricsResponse, err := st.ReadMetric(metricsRequest)
		if err != nil {
			log.Println("Request JSON store err", err,"--- request ---" , metricsRequest)
			c.Status(http.StatusNotFound)
			return
		}
		log.Println("Response JSON ----- ", metricsRequest)
		if key != "" {
			h := hmac.New(sha256.New, []byte(key))
			switch metricsResponse.MType {
			case "gauge":
				h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metricsResponse.ID, *metricsResponse.Value)))
				metricsResponse.Hash = hex.EncodeToString(h.Sum(nil))
			case "counter":
				h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricsResponse.ID, *metricsResponse.Delta)))
				metricsResponse.Hash = hex.EncodeToString(h.Sum(nil))
			}
		}
		c.JSON(http.StatusOK, metricsResponse)
	}
}

func UpdateMetricJSON(st storage.Storage, fs *storage.FileStorage, key string) gin.HandlerFunc {
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
		if key != "" {
			h := hmac.New(sha256.New, []byte(key))
			switch metricsRequest.MType {
			case "gauge":
				h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metricsRequest.ID, *metricsRequest.Value)))
			case "counter":
				h.Write([]byte(fmt.Sprintf("%s:counter:%d", metricsRequest.ID, *metricsRequest.Delta)))
			default:
				c.Status(http.StatusInternalServerError)
				return
			}
			if metricsRequest.Hash != hex.EncodeToString(h.Sum(nil)) {
				c.Status(http.StatusBadRequest)
				return
			}
		}
		err = st.InsertMetric(metricsRequest)
		if err != nil {
			log.Println("Insert metric err", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		if fs.Synchronize {
			err = fs.SaveStorageToFile(st)
			if err != nil {
				log.Println("Synchronized data saving was failed", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		c.Status(http.StatusOK)
	}
}

func PingDatabase(st storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := st.Ping()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

func RequestAllMetrics(st storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics, err := st.ReadAllMetrics()
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
	}
}

func BatchUpdateJSON(st storage.Storage, fs *storage.FileStorage, key string) gin.HandlerFunc {
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
				h := hmac.New(sha256.New, []byte(key))
				switch metric.MType {
				case "gauge":
					h.Write([]byte(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value)))
				case "counter":
					h.Write([]byte(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta)))
				}
				if metric.Hash != hex.EncodeToString(h.Sum(nil)) {
					continue
				}
				metricsBatchClean = append(metricsBatchClean, metric)
			}
			metricsBatch = metricsBatchClean
		}
		err = st.InsertBatchMetric(metricsBatch)
		if err != nil {
			log.Println("Error while update metrics from batch", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		if fs.Synchronize {
			err = fs.SaveStorageToFile(st)
			if err != nil {
				log.Println("Synchronized data saving was failed", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		c.Status(http.StatusOK)
	}
}

func WithoutID(c *gin.Context) {
	c.Status(http.StatusNotFound)
}
