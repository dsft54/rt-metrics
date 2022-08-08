// Package scheduller определяет планировщик для сборки данных работы системы.
package scheduller

import (
	"context"
	"sync"
	"time"

	"github.com/dsft54/rt-metrics/config/agent/settings"
)

// Scheduller состоит из двух тикеров poll и report и двух переменных состояния Pc и Rc. 
// А также логической переменной update, которая определяет необходимость очередного сбора метрик.
type Scheduller struct {
	poll   *time.Ticker
	rept   *time.Ticker
	Pc     *sync.Cond
	Rc     *sync.Cond
	Update bool
}

// NewScheduller функция-конструктор, которая формирует новый экземпляр Scheduller на основе конфигурации агента,
// а именно его PollInterval и ReportInterval.
func NewScheduller(cfg *settings.Config) *Scheduller {
	sch := new(Scheduller)
	sch.poll = time.NewTicker(cfg.PollInterval)
	sch.rept = time.NewTicker(cfg.ReportInterval)
	sch.Pc = sync.NewCond(&sync.Mutex{})
	sch.Rc = sync.NewCond(&sync.Mutex{})
	sch.Update = true
	return sch
}

// Start основная функция, определяющая работу планировщика. Ожидание срабатывания тикера и запускает бродкаст в 
// связанной с ним переменной состояния. Выход по context.Done, с вычитанием waitgroup.
func (sch *Scheduller) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-sch.poll.C:
			sch.Pc.Broadcast()
		case <-sch.rept.C:
			sch.Rc.Broadcast()
		case <-ctx.Done():
			return
		}
	}
}

// ExitRelease принудительный бродкаст по всем переменным состояния. Нужен для корректного выхода при завершении агента.
func (sch *Scheduller) ExitRelease() {
	sch.Pc.Broadcast()
	sch.Rc.Broadcast()
	sch.Update = false
}
