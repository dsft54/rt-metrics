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
	"fmt"
	"io/ioutil"
	"time"
)

type Config struct {
	Address        string        `env:"ADDRESS" json:"address"`
	HashKey        string        `env:"ADDRESS" json:"hash_key"`
	CryptoKey      string        `env:"CRYPTO_KEY" json:"crypto_key"`
	Config         string        `env:"CONFIG"`
	Batched        bool          `env:"BATCHED" json:"batched"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

func (c *Config) UnmarshalJSON(b []byte) error {
	type ConfigAlias Config
	aliasValue := &struct {
		*ConfigAlias
		PollInt   interface{} `json:"poll_interval"`
		ReportInt interface{} `json:"report_interval"`
	}{
		ConfigAlias: (*ConfigAlias)(c),
	}
	err := json.Unmarshal(b, &aliasValue)
	if err != nil {
		return err
	}
	switch value := aliasValue.PollInt.(type) {
	case float64:
		c.PollInterval = time.Duration(value)
	case string:
		c.PollInterval, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration: %#v", aliasValue.PollInt)
	}
	switch value := aliasValue.ReportInt.(type) {
	case float64:
		c.ReportInterval = time.Duration(value)
	case string:
		c.ReportInterval, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration: %#v", aliasValue.ReportInt)
	}
	return nil
}

func (c *Config) ParseFromFile() error {
	if c.Config == "" {
		return nil
	}
	data, err := ioutil.ReadFile(c.Config)
	if err != nil {
		return err
	}
	var fC Config
	err = json.Unmarshal(data, &fC)
	if err != nil {
		return err
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
	return nil
}
