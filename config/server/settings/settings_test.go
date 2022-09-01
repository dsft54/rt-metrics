// Package settings определяет структуру параметров сервера

package settings

import (
	"testing"
	"time"
)

func TestConfig_ParseFromFile(t *testing.T) {
	tests := []struct {
		name    string
		c       Config
		want    Config
		wantErr bool
	}{
		{
			name: "normal",
			c: Config{
				Address:     "1.1.1.1",
				StoreFile:   "test",
				HashKey:     "test",
				DatabaseDSN: "test",
				CryptoKey:   "test",
				Config:      "test_cfg.json",
				Restore:     false,
			},
			want: Config{
				Address:     "1.1.1.1",
				StoreFile:   "test",
				HashKey:     "test",
				DatabaseDSN: "test",
				CryptoKey:   "test",
				Config:      "test_cfg.json",
				Restore:     false,
			},
			wantErr: false,
		},
		{
			name: "return ''",
			c: Config{
				Address:     "1.1.1.1",
				StoreFile:   "test",
				HashKey:     "test",
				DatabaseDSN: "test",
				CryptoKey:   "test",
				Config:      "",
				Restore:     false,
			},
			want: Config{
				Address:     "1.1.1.1",
				StoreFile:   "test",
				HashKey:     "test",
				DatabaseDSN: "test",
				CryptoKey:   "test",
				Config:      "",
				Restore:     false,
			},
			wantErr: false,
		},
		{
			name: "return error",
			c: Config{
				Address:     "1.1.1.1",
				StoreFile:   "test",
				HashKey:     "test",
				DatabaseDSN: "test",
				CryptoKey:   "test",
				Config:      "return",
				Restore:     false,
			},
			want: Config{
				Address:     "1.1.1.1",
				StoreFile:   "test",
				HashKey:     "test",
				DatabaseDSN: "test",
				CryptoKey:   "test",
				Config:      "return",
				Restore:     false,
			},
			wantErr: true,
		},
		{
			name: "correct",
			c: Config{
				Address:       "",
				StoreFile:     "",
				HashKey:       "",
				DatabaseDSN:   "",
				CryptoKey:     "",
				StoreInterval: 0 * time.Second,
				Config:        "test_cfg.json",
				Restore:       false,
			},
			want: Config{
				Address:       "localhost:8080",
				StoreFile:     "/path/to/file.db",
				HashKey:       "test",
				DatabaseDSN:   "postgres",
				CryptoKey:     "/path/to/key.pem",
				StoreInterval: 1 * time.Second,
				Config:        "test_cfg.json",
				Restore:       true,
			},
			wantErr: false,
		},
		{
			name: "unmarshall err",
			c: Config{
				Address:     "",
				StoreFile:   "",
				HashKey:     "",
				DatabaseDSN: "",
				CryptoKey:   "",
				Config:      "test_cfg_fail.json",
				Restore:     false,
			},
			want:    Config{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.ParseFromFile(); (err != nil) != tt.wantErr {
				t.Errorf("Config.ParseFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		c       *Config
		json    string
		wantErr bool
	}{
		{
			name: "Normal string test",
			c:    &Config{},
			json: `{
                "address": "localhost:8080",
                "restore": true,
                "store_file": "21",
                "database_dsn": "postgres",
                "store_interval": "2s",
                "crypto_key": "/path/to/key.pem"
            }`,
			wantErr: false,
		},
		{
			name:    "Normal float test",
			c:       &Config{},
			json: `{
                "address": "localhost:8080",
                "restore": true,
                "store_file": "21",
                "database_dsn": "postgres",
                "store_interval": 2,
                "crypto_key": "/path/to/key.pem"
            }`,
			wantErr: false,
		},
		{
			name:    "String parse error test",
			c:       &Config{},
			json: `{
                "address": "localhost:8080",
                "restore": true,
                "store_file": "21",
                "database_dsn": "postgres",
                "store_interval": "2g",
                "crypto_key": "/path/to/key.pem"
            }`,
			wantErr: true,
		},
		{
			name:    "Unmarshall error test",
			c:       &Config{},
			json: `{
                "address": "localhost:8080",
                "restore": true,
                "store_file": 21,
                "database_dsn": "postgres",
                "store_interval": "2s",
                "crypto_key": "/path/to/key.pem"
            }`,
			wantErr: true,
		},
		{
			name:    "Unmarshall type error test",
			c:       &Config{},
			json: `{
                "address": "localhost:8080",
                "restore": true,
                "store_file": "21",
                "database_dsn": "postgres",
                "store_interval": false,
                "crypto_key": "/path/to/key.pem"
            }`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.UnmarshalJSON([]byte(tt.json)); (err != nil) != tt.wantErr {
				t.Errorf("Config.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
