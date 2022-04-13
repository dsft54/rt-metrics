package settings

import "time"

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
}

var Cfg Config