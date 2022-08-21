// Package settings агента определяет основные настройки запуска агента при помощи
// переменных окружения или флагов.
// Address - определяет адрес сервера для отправки метрик.
// PollInterval - частота сбора метрик агента в секндах.
// ReportInterval - частота отправки метрик на сервер в секундах.
// HashKey - ключ для подписи хеша.
// Batched - отправлять метрики списком или штучно.
package settings

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type Config struct {
	Address        string        `env:"ADDRESS" json:"address"`
	HashKey        string        `env:"ADDRESS" json:"hash_key"`
	CryptoKey      string        `env:"CRYPTO_KEY" json:"crypto_key"`
	Config         string        `env:"CONFIG"`
	Batched        bool          `env:"BATCHED" json:"batched"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" json:"report_interval"`
}

func (c *Config) ParseFromFile() {
	if c.Config == "" {
		return
	}
	data, err := ioutil.ReadFile(c.Config)
	if err != nil {
		return
	}
	var fC Config
	err = json.Unmarshal(data, &fC)
	if err != nil {
		return
	}
	if c.Address == "" && fC.Address != "" {
		c.Address = fC.Address
	}
	if c.HashKey == "" && fC.HashKey != "" {
		c.HashKey = fC.HashKey
	}
	if c.CryptoKey == "" && fC.CryptoKey != "" {
		c.CryptoKey = fC.CryptoKey
	}
	if c.PollInterval == 0 && fC.PollInterval != 0 {
		c.PollInterval = fC.PollInterval
	}
	if c.ReportInterval == 0 && fC.ReportInterval != 0 {
		c.ReportInterval = fC.ReportInterval
	}
}
