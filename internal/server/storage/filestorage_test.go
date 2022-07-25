package storage

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/dsft54/rt-metrics/config/server/settings"
)

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name string
		cfg  settings.Config
		want *FileStorage
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
	tf, err := ioutil.TempFile("./", "test")
	defer tf.Close()
	if err != nil {
		t.Error(err)
	}
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
		})
	}
	
}
