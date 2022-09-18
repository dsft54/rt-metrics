package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/dsft54/rt-metrics/config/server/settings"
	"github.com/dsft54/rt-metrics/internal/server/storage"
)

func Test_initStorages(t *testing.T) {
	tests := []struct {
		want1  *storage.FileStorage
		want   storage.IStorage
		ctx    context.Context
		name   string
		config settings.Config
	}{
		{
			name: "normal mem init",
			ctx:  context.Background(),
			config: settings.Config{
				DatabaseDSN: "",
				StoreFile:   "",
			},
			want: &storage.MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			want1: storage.NewFileStorage(settings.Config{
				DatabaseDSN: "",
				StoreFile:   "",
			}),
		},
		{
			name: "normal db err init",
			ctx:  context.Background(),
			config: settings.Config{
				DatabaseDSN: "error",
				StoreFile:   "",
			},
			want: &storage.MemoryStorage{
				GaugeMetrics:   map[string]float64{},
				CounterMetrics: map[string]int64{},
			},
			want1: storage.NewFileStorage(settings.Config{
				DatabaseDSN: "",
				StoreFile:   "",
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := initStorages(tt.ctx, tt.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initStorages() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("initStorages() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_setupGinRouter(t *testing.T) {
	tests := []struct {
		st   storage.IStorage
		fs   *storage.FileStorage
		kp   string
		an   string
		name string
		wantErr bool
	}{
		{
			name: "normal gin router",
			st:   &storage.MemoryStorage{},
			fs:   &storage.FileStorage{},
			wantErr: false,
		},
		{
			name: "addons key gin router",
			st:   &storage.MemoryStorage{},
			fs:   &storage.FileStorage{},
			kp: "err",
			wantErr: true,
		},
		{
			name: "addons key gin router",
			st:   &storage.MemoryStorage{},
			fs:   &storage.FileStorage{},
			kp: "test",
			wantErr: false,
		},
		{
			name: "addons pub key gin router",
			st:   &storage.MemoryStorage{},
			fs:   &storage.FileStorage{},
			kp: "teste",
			wantErr: true,
		},
		{
			name: "addons an gin router",
			st:   &storage.MemoryStorage{},
			fs:   &storage.FileStorage{},
			an: "172.16.0.0/16",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := setupGinRouter(tt.st, tt.fs, tt.kp, tt.an)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
