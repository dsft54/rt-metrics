package scheduller

import (
	"context"
	"sync"
	"time"

	"github.com/dsft54/rt-metrics/cmd/agent/settings"
)

type Scheduller struct {
	poll   *time.Ticker
	rept   *time.Ticker
	Pc     *sync.Cond
	Rc     *sync.Cond
	Update bool
}

func NewScheduller(cfg *settings.Config) *Scheduller {
	sch := new(Scheduller)
	sch.poll = time.NewTicker(cfg.PollInterval)
	sch.rept = time.NewTicker(cfg.ReportInterval)
	sch.Pc = sync.NewCond(&sync.Mutex{})
	sch.Rc = sync.NewCond(&sync.Mutex{})
	sch.Update = true
	return sch
}

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

func (sch *Scheduller) ExitRelease() {
	sch.Pc.Broadcast()
	sch.Rc.Broadcast()
	sch.Update = false
}
