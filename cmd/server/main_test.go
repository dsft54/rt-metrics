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
		name string
	}{
		{
			name: "normal gin router",
			st:   &storage.MemoryStorage{},
			fs:   &storage.FileStorage{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setupGinRouter(tt.st, tt.fs, "")
			if len(got.Handlers) != 4 {
				t.Error("Failed to build middleware chain, should be: ", len(got.Handlers))
			}
			if len(got.RouterGroup.Handlers) != 4 {
				t.Error("Failed to build routergroup chain, should be: ", len(got.RouterGroup.Handlers))
			}

		})
	}
}
