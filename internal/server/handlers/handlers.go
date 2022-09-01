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
	
	"github.com/dsft54/rt-metrics/internal/server/storage"
)

// ParametersUpdate используется для обработки POST запроса для обновления/записи
// метрики с использованием параметров в url запроса в формате "/update/:type/:name/:value".
// Если требуется синхронная запись в файл, она осуществляется через метод FileStorage.
func ParametersUpdate(st storage.IStorage, fs *storage.FileStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		mType := c.Param("type")
		mName := c.Param("name")
		mValue := c.Param("value")
		code, err := st.ParamsUpdate(mType, mName, mValue)
		if err != nil {
			log.Println(err, "Type:", mType, "Name:", mName, "Value:", mValue, "Code", code)
		}
		if code == 200 && fs.Synchronize {
			err := fs.SaveStorageToFile(st)
			if err != nil {
				log.Println("Synchronized data saving was failed", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		c.Status(code)
	}
}

// AddressedRequest используется для обработки GET запроса на получение одной метрики,
// тип и название которой будет получено из параметров url запроса вида "/value/:type/:name".
func AddressedRequest(st storage.IStorage) gin.HandlerFunc {
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
	}
}

// RequestMetricJSON используется для обработки POST запроса на получения одной метрики,
// где тип и название метрики передается в теле запроса в формате json по url /value/.
func RequestMetricJSON(st storage.IStorage, key string) gin.HandlerFunc {
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
		metricsResponse, err := st.ReadMetric(metricsRequest)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
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

// UpdateMetricJSON используется для обработки POST запроса на обновление/запись одной метрики,
// где тип, название и значение метрики передается в теле запроса в формате json по url /update/.
// В случае если при запуске сервера был указан ключ, считается хеш полученной метрики и сравнивается
// с тем, который был получен от агента. При неравенстве хешей такой запрос отбрасывается.
// При необходимости запись дублируется в файл.
func UpdateMetricJSON(st storage.IStorage, fs *storage.FileStorage, key string) gin.HandlerFunc {
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

// PingDatabase обработчик GET запросов, который позволяет проверить состояние подключения к базе данных по пути /ping/,
// если активное хранилище развенуто в памяти, вернет ошибку.
func PingDatabase(st storage.IStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := st.Ping()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
}

// RequestAllMetrics возвращает значения всех сохраненных метрик в json формате. Предназначен для обработки GET запроса
// на /.
func RequestAllMetrics(st storage.IStorage) gin.HandlerFunc {
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

// BatchUpdateJSON предназначен для обновления списка метрик полученных в теле POST запроса
// в формате json. Также проверяется хеш при наличии ключа. При необходимости запись дублируется в файл.
func BatchUpdateJSON(st storage.IStorage, fs *storage.FileStorage, key string) gin.HandlerFunc {
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

// WithoutID возвращает ошибку 404 при попытке сделать POST запрос на "/update/counter/" и "/update/value/"
// т.е. без указания названия искомой метрики.
func WithoutID(c *gin.Context) {
	c.Status(http.StatusNotFound)
}
