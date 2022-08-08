package storage

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dsft54/rt-metrics/config/server/settings"
	"github.com/go-playground/assert"
)

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		want *FileStorage
		name string
		cfg  settings.Config
	}{
		{
			name: "No path, no restore",
			cfg: settings.Config{
				StoreInterval: 0,
				StoreFile:     "",
				Restore:       false,
			},
			want: &FileStorage{
				FilePath:    "",
				StoreData:   false,
				Synchronize: false,
			},
		},
		{
			name: "No path, restore",
			cfg: settings.Config{
				StoreInterval: 0,
				StoreFile:     "",
				Restore:       true,
			},
			want: &FileStorage{
				FilePath:    "",
				StoreData:   false,
				Synchronize: false,
			},
		},
		{
			name: "Path, restore, sync",
			cfg: settings.Config{
				StoreInterval: 0,
				StoreFile:     "test",
				Restore:       true,
			},
			want: &FileStorage{
				FilePath:    "test",
				StoreData:   true,
				Synchronize: true,
			},
		},
		{
			name: "Path, do not restore, do sync with interval",
			cfg: settings.Config{
				StoreInterval: 10,
				StoreFile:     "test",
				Restore:       false,
			},
			want: &FileStorage{
				FilePath:    "test",
				StoreData:   true,
				Synchronize: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileStorage(tt.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStorage_OpenToWrite(t *testing.T) {
	tests := []struct {
		name    string
		f       *FileStorage
		path    string
		wantErr bool
	}{
		{
			name:    "File exists",
			f:       &FileStorage{},
			path:    "test",
			wantErr: false,
		},
		{
			name:    "File do not exists or empty",
			f:       &FileStorage{},
			path:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.OpenToWrite(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("FileStorage.OpenToWrite() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.path != "" {
				err := tt.f.File.Close()
				if err != nil {
					t.Error(err)
				}
				err = os.Remove(tt.f.File.Name())
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestFileStorage_SaveStorageToFile(t *testing.T) {
	tests := []struct {
		name       string
		f          *FileStorage
		s          IStorage
		wantInFile string
		wantErr    bool
	}{
		{
			name: "Save memstorage",
			f: &FileStorage{
				FilePath: "test",
			},
			s: &MemoryStorage{
				GaugeMetrics: map[string]float64{
					"Alloc": 3.14,
				},
				CounterMetrics: map[string]int64{
					"Counter": 3,
				},
			},
			wantInFile: "[{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":3.14},{\"id\":\"Counter\",\"type\":\"counter\",\"delta\":3}]",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.SaveStorageToFile(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("FileStorage.SaveStorageToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			data, err := ioutil.ReadFile(tt.f.FilePath)
			if err != nil {
				t.Error(err)
			}
			if tt.f.FilePath != "" {
				err = os.Remove(tt.f.FilePath)
				if err != nil {
					t.Error(err)
				}
			}
			assert.Equal(t, tt.wantInFile, string(data))
		})
	}
}

func TestFileStorage_IntervalUpdate(t *testing.T) {
	tests := []struct {
		f    *FileStorage
		dur  time.Duration
		ctx  context.Context
		s    IStorage
		name string
	}{
		{
			name: "context exit",
			f: NewFileStorage(settings.Config{
				StoreFile: "test",
			}),
			dur: 500 * time.Millisecond,
			s: &MemoryStorage{
				GaugeMetrics:   map[string]float64{"Alloc": 3.14},
				CounterMetrics: map[string]int64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				cancel context.CancelFunc
				err    error
			)
			tt.ctx, cancel = context.WithCancel(context.Background())
			tt.f.File, err = os.OpenFile(tt.f.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				t.Error(err)
			}
			go tt.f.IntervalUpdate(tt.ctx, tt.dur, tt.s)
			<-time.NewTimer(700 * time.Millisecond).C
			cancel()
			data, err := ioutil.ReadFile("test")
			if err != nil {
				t.Error("File was not created")
			}
			if string(data) != "[{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":3.14}]" {
				t.Error("Data in file not correct", string(data))
			}
			err = tt.f.File.Close()
			if err != nil {
				t.Error(err)
			}
			err = os.Remove(tt.f.FilePath)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
