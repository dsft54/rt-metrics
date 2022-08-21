// Package settings агента определяет основные настройки запуска агента при помощи
// переменных окружения или флагов.
// Address - определяет адрес сервера для отправки метрик.
// PollInterval - частота сбора метрик агента в секндах.
// ReportInterval - частота отправки метрик на сервер в секундах.
// HashKey - ключ для подписи хеша.
// Batched - отправлять метрики списком или штучно.
package settings

import "time"

type Config struct {
	Address        string        `env:"ADDRESS"`
	HashKey        string        `env:"ADDRESS"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
	Batched        bool          `env:"BATCHED"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}
