package storage

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/dsft54/rt-metrics/cmd/server/settings"
)

type FileStorage struct {
	File        *os.File
	FilePath    string
	StoreData   bool
	Synchronize bool
}

func NewFileStorage(cfg settings.Config) *FileStorage {
	fs := new(FileStorage)
	if cfg.StoreFile == "" {
		fs.StoreData = false
		fs.Synchronize = false
	} else {
		fs.FilePath = cfg.StoreFile
		fs.StoreData = true
	}
	if cfg.StoreInterval == 0 {
		fs.Synchronize = true
	}
	if cfg.StoreInterval > 0 {
		fs.Synchronize = false
	}
	return fs
}


func (f *FileStorage) OpenToWrite(path string) (err error) {
	f.File, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) SaveMemDataToFile(condition bool, m *MemoryStorage) error {
	if !condition {
		return nil
	}
	err := f.OpenToWrite(f.FilePath)
	if err != nil {
		return err
	}
	m.WriteMetricsToFile(f.File)
	f.File.Close()
	return nil
}

func (f *FileStorage) SaveDBDataToFile(condition bool, d *DBStorage) error {
	if !condition {
		return nil
	}
	metrics, err := d.DBReadAll()
	if err != nil {
		return err
	}
	log.Println("Tried to save db to file on exit", metrics)
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	err = f.OpenToWrite(f.FilePath)
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

func (f *FileStorage) IntervalUpdateMem(ctx context.Context, dur time.Duration, s *MemoryStorage) {
	intervalTicker := time.NewTicker(dur)
	for {
		select {
		case <-intervalTicker.C:
			f.OpenToWrite(f.FilePath)
			s.WriteMetricsToFile(f.File)
			f.File.Close()
		case <-ctx.Done():
			return
		}
	}
}

func (f *FileStorage) IntervalUpdateDB(ctx context.Context, dur time.Duration, db *DBStorage) {
	intervalTicker := time.NewTicker(dur)
	for {
		select {
		case <-intervalTicker.C:
			err := db.DBSaveToFile(f)
			if err != nil {
				log.Print("Failed to save metrics from db to file with ticker", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
