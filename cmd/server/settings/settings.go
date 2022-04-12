package settings

import "time"

type Config struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"5s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

var Cfg Config
