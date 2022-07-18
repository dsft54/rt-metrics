package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/jackc/pgx"
)

type DBStorage struct {
	Connection *pgx.Conn
	ConnConfig pgx.ConnConfig
	Context    context.Context
	TableName  string
}

func (d *DBStorage) InsertMetric(m *Metrics) error {
	switch m.MType {
	case "gauge":
		_, err := d.Connection.Exec(
			`INSERT INTO rt_metrics (id, mtype, value, hash)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id) DO UPDATE
				SET value = excluded.value, hash = excluded.hash;`,
			m.ID, m.MType, m.Value, m.Hash)
		if err != nil {
			return err
		}
	case "counter":
		_, err := d.Connection.Exec(
			`INSERT INTO rt_metrics (id, mtype, delta, hash)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id) DO UPDATE
				SET delta = excluded.delta + rt_metrics.delta, hash = excluded.hash;`,
			m.ID, m.MType, m.Delta, m.Hash)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DBStorage) ReadAllMetrics() ([]Metrics, error) {
	// Read all metrics from db
	var metricsSlice []Metrics
	rows, err := d.Connection.Query("SELECT * FROM rt_metrics;")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		metric := Metrics{}
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value, &metric.Hash)
		if err != nil {
			return nil, err
		}
		metricsSlice = append(metricsSlice, metric)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return metricsSlice, nil
}

func (d *DBStorage) ParamsUpdate(metricType, metricID, metricValue string) (int, error) {
	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return 400, err
		}
		_, err = d.Connection.Exec(
			`INSERT INTO rt_metrics (id, mtype, value)
				VALUES ($1, $2, $3)
				ON CONFLICT (id) DO UPDATE
				SET value = excluded.value;`,
			metricID, metricType, value)
		if err != nil {
			return 400, err
		}
		return 200, nil
	case "counter":
		delta, err := strconv.Atoi(metricValue)
		if err != nil {
			return 400, err
		}
		_, err = d.Connection.Exec(
			`INSERT INTO rt_metrics (id, mtype, delta)
				VALUES ($1, $2, $3)
				ON CONFLICT (id) DO UPDATE
				SET delta = excluded.delta + rt_metrics.delta;`,
			metricID, metricType, delta)
		if err != nil {
			return 400, err
		}
		return 200, nil
	}
	return 501, errors.New("Wrong metric type - " + metricType)
}

func (d *DBStorage) SaveToFile(f *os.File) error {
	metrics, err := d.ReadAllMetrics()
	if err != nil {
		return err
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) UploadFromFile(path string) error {
	var metricsSlice []Metrics
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &metricsSlice)
	if err != nil {
		return err
	}
	for _, metric := range metricsSlice {
		switch metric.MType {
		case "gauge":
			err := d.InsertMetric(&metric)
			if err != nil {
				return err
			}
		case "counter":
			_, err := d.Connection.Exec(
				`INSERT INTO rt_metrics (id, mtype, delta, hash)
						VALUES ($1, $2, $3, $4)
						ON CONFLICT (id) DO UPDATE
						SET delta = excluded.delta, hash = excluded.hash;`,
				metric.ID, metric.MType, metric.Delta, metric.Hash)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DBStorage) ReadMetric(rm *Metrics) (*Metrics, error) {
	// Read specific metric from db
	row := d.Connection.QueryRow(
		`SELECT delta, value, hash FROM rt_metrics 
			WHERE rt_metrics.id = $1 AND rt_metrics.mtype = $2`,
		rm.ID, rm.MType)
	err := row.Scan(&rm.Delta, &rm.Value, &rm.Hash)
	if err != nil {
		return nil, err
	}
	return rm, nil
}

func (d *DBStorage) InsertBatchMetric(metrics []Metrics) error {
	for _, metric := range metrics {
		err := d.InsertMetric(&metric)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DBStorage) DBFlushTable() error {
	// Empty table
	_, err := d.Connection.Exec("TRUNCATE rt_metrics;")
	if err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) DBSaveToFile(f *FileStorage) error {
	err := f.OpenToWrite(f.FilePath)
	if err != nil {
		return err
	}
	metrics, err := d.ReadAllMetrics()
	if err != nil {
		return err
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	_, err = f.File.Write(data)
	if err != nil {
		return err
	}
	f.File.Close()
	return nil
}


func (d *DBStorage) Ping() error {
	err := d.Connection.Ping(d.Context)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) DBConnectStorage(ctx context.Context, auth, tableName string) error {
	var err error
	if auth == "" {
		return nil
	}
	d.ConnConfig, err = pgx.ParseConnectionString(auth)
	if err != nil {
		return errors.New("DB auth uri parse failed")
	}
	d.Connection, err = pgx.Connect(d.ConnConfig)
	if err != nil {
		return errors.New("WARNING! DB connection failed")
	}
	d.TableName = tableName
	d.Context = ctx
	return nil
}

func (d *DBStorage) DBCheckTableExists() error {
	// Check if table exists
	row := d.Connection.QueryRow(
		`SELECT EXISTS (
			SELECT FROM pg_tables
			WHERE schemaname = 'public'
			AND tablename  = rt_metrics
		);`)
	var tableExists bool
	err := row.Scan(&tableExists)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) DBCreateTable() error {
	// Create table
	_, err := d.Connection.Exec(
		`CREATE TABLE IF NOT EXISTS rt_metrics (
			id TEXT UNIQUE,
			mtype TEXT,
			delta BIGINT,
			value DOUBLE PRECISION,
			hash TEXT
		);`)
	if err != nil {
		return err
	}
	return nil
}
