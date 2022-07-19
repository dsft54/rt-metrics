package storage

import (
	"context"
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

func (f *FileStorage) SaveStorageToFile(s Storage) error {
	err := f.OpenToWrite(f.FilePath)
	if err != nil {
		return err
	}
	s.SaveToFile(f.File)
	f.File.Close()
	return nil
}

func (f *FileStorage) IntervalUpdate(ctx context.Context, dur time.Duration, s Storage) {
	intervalTicker := time.NewTicker(dur)
	for {
		select {
		case <-intervalTicker.C:
			f.OpenToWrite(f.FilePath)
			s.SaveToFile(f.File)
			f.File.Close()
		case <-ctx.Done():
			return
		}
	}
}