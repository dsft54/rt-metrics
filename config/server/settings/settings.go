// Модуль setting определяет структуру параметров сервера
package settings

import "time"

// Config - структура описывающая основные настройки сервера. Принимает переменные окружения и флаги. 
type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	HashKey       string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}
