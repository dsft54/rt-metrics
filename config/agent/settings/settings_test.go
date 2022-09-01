package settings

import (
	"testing"
	"time"
)

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
				"hash_key": "test",
				"crypto_key": "/path/to/key.pem",
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": "1s",
				"report_interval": "1s"
			}`,
			wantErr: false,
		},
		{
			name: "Normal float test",
			c:    &Config{},
			json: `{
				"address": "localhost:8080",
				"hash_key": "test",
				"crypto_key": "/path/to/key.pem",
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": 1,
				"report_interval": 1
			}`,
			wantErr: false,
		},
		{
			name: "String parse error test",
			c:    &Config{},
			json: `{
				"address": "localhost:8080",
				"hash_key": "test",
				"crypto_key": "/path/to/key.pem",
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": "1g",
				"report_interval": "1s"
			}`,
			wantErr: true,
		},
		{
			name: "String parse error test 2",
			c:    &Config{},
			json: `{
				"address": "localhost:8080",
				"hash_key": "test",
				"crypto_key": "/path/to/key.pem",
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": "1s",
				"report_interval": "2g"
			}`,
			wantErr: true,
		},
		{
			name: "Unmarshall error test",
			c:    &Config{},
			json: `{
				"address": "localhost:8080",
				"hash_key": "test",
				"crypto_key": 22,
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": "1s",
				"report_interval": "1s"
			}`,
			wantErr: true,
		},
		{
			name: "Unmarshall type error test",
			c:    &Config{},
			json: `{
				"address": "localhost:8080",
				"hash_key": "test",
				"crypto_key": "/path/to/key.pem",
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": false,
				"report_interval": false
			}`,
			wantErr: true,
		},
		{
			name: "Unmarshall type error test",
			c:    &Config{},
			json: `{
				"address": "localhost:8080",
				"hash_key": "test",
				"crypto_key": "/path/to/key.pem",
				"store_file": "/path/to/file.db",
				"batched": false,
				"poll_interval": 2,
				"report_interval": false
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
				Address:   "1.1.1.1",
				HashKey:   "test",
				CryptoKey: "test",
				Config:    "test_cfg.json",
			},
			want: Config{
				Address:   "1.1.1.1",
				HashKey:   "test",
				CryptoKey: "test",
				Config:    "test_cfg.json",
			},
			wantErr: false,
		},
		{
			name: "return ''",
			c: Config{
				Address:   "1.1.1.1",
				HashKey:   "test",
				CryptoKey: "test",
				Config:    "",
			},
			want: Config{
				Address:   "1.1.1.1",
				HashKey:   "test",
				CryptoKey: "test",
				Config:    "",
			},
			wantErr: false,
		},
		{
			name: "return error",
			c: Config{
				Address:   "1.1.1.1",
				HashKey:   "test",
				CryptoKey: "test",
				Config:    "return",
			},
			want: Config{
				Address:   "1.1.1.1",
				HashKey:   "test",
				CryptoKey: "test",
				Config:    "return",
			},
			wantErr: true,
		},
		{
			name: "correct",
			c: Config{
				Address:        "",
				HashKey:        "",
				CryptoKey:      "",
				PollInterval:   0 * time.Second,
				ReportInterval: 0 * time.Second,
				Config:         "test_cfg.json",
			},
			want: Config{
				Address:       "localhost:8080",
				HashKey:       "test",
				CryptoKey:     "/path/to/key.pem",
				PollInterval: 1 * time.Second,
				ReportInterval: 1 * time.Second,
				Config:        "test_cfg.json",
			},
			wantErr: false,
		},
		{
			name: "unmarshall err",
			c: Config{
				Address:     "",
				HashKey:     "",
				CryptoKey:   "",
				Config:      "test_cfg_fail.json",
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
