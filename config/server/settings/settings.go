// Package settings определяет структуру параметров сервера
package settings

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Config - структура описывающая основные настройки сервера. Принимает переменные окружения и флаги.
type Config struct {
	Address       string        `env:"ADDRESS" json:"address"`
	StoreFile     string        `env:"STORE_FILE" json:"store_file"`
	HashKey       string        `env:"KEY" json:"hash_key"`
	DatabaseDSN   string        `env:"DATABASE_DSN" json:"database_dsn"`
	CryptoKey     string        `env:"CRYPTO_KEY" json:"crypto_key"`
	Config        string        `env:"CONFIG"`
	Restore       bool          `env:"RESTORE" json:"restore"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" json:"store_interval"`
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
	if c.StoreFile == "" && fC.StoreFile != "" {
		c.StoreFile = fC.StoreFile
	}
	if c.HashKey == "" && fC.HashKey != "" {
		c.HashKey = fC.HashKey
	}
	if c.DatabaseDSN == "" && fC.DatabaseDSN != "" {
		c.DatabaseDSN = fC.DatabaseDSN
	}
	if c.CryptoKey == "" && fC.CryptoKey != "" {
		c.CryptoKey = fC.CryptoKey
	}
	if c.StoreInterval == 0 && fC.StoreInterval != 0 {
		c.StoreInterval = fC.StoreInterval
	}
}