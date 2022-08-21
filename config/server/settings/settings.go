// Package settings определяет структуру параметров сервера
package settings

import "time"

// Config - структура описывающая основные настройки сервера. Принимает переменные окружения и флаги.
type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreFile     string        `env:"STORE_FILE"`
	HashKey       string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
	CryptoKey     string        `env:"CRYPTO_KEY"`
	Restore       bool          `env:"RESTORE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
}
