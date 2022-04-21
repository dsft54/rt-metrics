package settings

import "time"

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	HashKey       string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}
