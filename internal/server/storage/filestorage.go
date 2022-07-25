package storage

import (
	"context"
	"os"
	"time"

	"github.com/dsft54/rt-metrics/config/server/settings"
)

// FileStorage стуктура, описывающая файл куда/откуда будут сохранены/загружены метрики, путь до него и
// логические переменные, определяющие необходимость загрузки метрик в хранилище при старте сервера
// и сохранении их при остановке - StoreData, и определяющие необходимость синхронной записи метрик не только
// в хранилище, но и в файл - Synchronize.
type FileStorage struct {
	File        *os.File
	FilePath    string
	StoreData   bool
	Synchronize bool
}

// NewFileStorage функция-конструктор для структуры FileStorage. В зависимости от конфигурации запуска сервера,
// принимает (если он есть) путь до файла, и получает значения логических переменных StoreData и Synchronize.
func NewFileStorage(cfg settings.Config) *FileStorage {
	fs := new(FileStorage)
	if cfg.StoreFile == "" {
		fs.StoreData = false
		fs.Synchronize = false
		return fs
	}
	fs.FilePath = cfg.StoreFile
	fs.StoreData = true
	if cfg.StoreInterval == 0 {
		fs.Synchronize = true
	}
	if cfg.StoreInterval > 0 {
		fs.Synchronize = false
	}
	return fs
}

// OpenToWrite открывает файл для записи по указанному пути.
func (f *FileStorage) OpenToWrite(path string) (err error) {
	f.File, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	return nil
}

// SaveStorageToFile сохраняет текущий активный storage в файл.
func (f *FileStorage) SaveStorageToFile(s IStorage) error {
	err := f.OpenToWrite(f.FilePath)
	if err != nil {
		return err
	}
	s.SaveToFile(f.File)
	f.File.Close()
	return nil
}

// IntervalUpdate создает тикер и в бесконечном цикле ожидает либо срабатывания тикера для того,
// чтобы сохранить текущий storage в файл, либо ctx.Done, для того, чтобы завершить срабатывание цикла.
func (f *FileStorage) IntervalUpdate(ctx context.Context, dur time.Duration, s IStorage) {
	intervalTicker := time.NewTicker(dur)
	for {
		select {
		case <-intervalTicker.C:
			s.SaveToFile(f.File)
		case <-ctx.Done():
			return
		}
	}
}
