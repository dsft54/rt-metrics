package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dsft54/rt-metrics/config/agent/settings"
	"github.com/dsft54/rt-metrics/internal/agent/scheduller"
	"github.com/dsft54/rt-metrics/internal/agent/storage"
	"github.com/go-resty/resty/v2"
)

func Test_sendData(t *testing.T) {
	client := resty.New()
	type args struct {
		url     string
		keyPath string
		metrics interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "server offline (or not correct)",
			args: args{
				url: "http://localhost:808",
				metrics: storage.Metrics{
					ID:    "",
					MType: "",
				},
				keyPath: "",
			},
			wantErr: true,
		},
		{
			name: "keypath",
			args: args{
				url: "http://localhost:808",
				metrics: storage.Metrics{
					ID:    "",
					MType: "",
				},
				keyPath: "test.pub",
			},
			wantErr: true,
		},
		{
			name: "keypath err",
			args: args{
				url: "http://localhost:808",
				metrics: storage.Metrics{
					ID:    "",
					MType: "",
				},
				keyPath: "t",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendData(tt.args.url, tt.args.keyPath, &tt.args.metrics, client); (err != nil) != tt.wantErr {
				t.Errorf("sendData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportMetrics(t *testing.T) {
	tests := []struct {
		ctx  context.Context
		sch  *scheduller.Scheduller
		cfg  *settings.Config
		s    *storage.MemStorage
		wg   *sync.WaitGroup
		name string
	}{
		{
			name: "context exit",
			sch: scheduller.NewScheduller(&settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			}),
			cfg: &settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			},
			s:  storage.NewMemStorage(),
			wg: new(sync.WaitGroup),
		},
		{
			name: "do not update",
			sch: scheduller.NewScheduller(&settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			}),
			cfg: &settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			},
			s:  storage.NewMemStorage(),
			wg: new(sync.WaitGroup),
		},
		{
			name: "normal",
			sch: scheduller.NewScheduller(&settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			}),
			cfg: &settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			},
			s:  storage.NewMemStorage(),
			wg: new(sync.WaitGroup),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "context exit":
				var cancel context.CancelFunc
				tt.wg.Add(1)
				tt.ctx, cancel = context.WithCancel(context.Background())
				go reportMetrics(tt.ctx, tt.sch, tt.cfg, tt.s, tt.wg)
				cancel()
				select {
				case <-time.NewTimer(500 * time.Millisecond).C:
					t.Error("Goroutine timeout error")
				case <-wrapWait(tt.wg):
					t.Log("Ok")
				}
			case "do not update":
				tt.ctx = context.Background()
				tt.wg.Add(1)
				tt.sch.Update = false
				go reportMetrics(tt.ctx, tt.sch, tt.cfg, tt.s, tt.wg)
				<-time.NewTimer(500 * time.Millisecond).C
				tt.sch.Rc.Broadcast()
				select {
				case <-time.NewTimer(500 * time.Millisecond).C:
					t.Error("Goroutine timeout error")
				case <-wrapWait(tt.wg):
					t.Log("Ok")
				}
			case "normal":
				// just should be stdout
				var cancel context.CancelFunc
				tt.wg.Add(1)
				tt.ctx, cancel = context.WithCancel(context.Background())
				go reportMetrics(tt.ctx, tt.sch, tt.cfg, tt.s, tt.wg)
				<-time.NewTimer(1000 * time.Millisecond).C
				tt.sch.Rc.Broadcast()
				cancel()
				select {
				case <-time.NewTimer(3000 * time.Millisecond).C:
					t.Error("Goroutine timeout error")
				case <-wrapWait(tt.wg):
					t.Log("Ok")
				}
			}
		})
	}
}

// helper function to allow using WaitGroup in a select
func wrapWait(wg *sync.WaitGroup) <-chan struct{} {
	out := make(chan struct{})
	go func() {
		wg.Wait()
		out <- struct{}{}
	}()
	return out
}

func Test_pollRuntimeMetrics(t *testing.T) {
	tests := []struct {
		ctx  context.Context
		c    *sync.Cond
		s    *storage.MemStorage
		wg   *sync.WaitGroup
		name string
	}{
		{
			name: "context exit, data collected",
			c:    sync.NewCond(&sync.Mutex{}),
			s:    storage.NewMemStorage(),
			wg:   new(sync.WaitGroup),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cancel context.CancelFunc
			tt.wg.Add(1)
			tt.ctx, cancel = context.WithCancel(context.Background())
			go pollRuntimeMetrics(tt.ctx, tt.c, tt.s, tt.wg)
			<-time.NewTimer(100 * time.Millisecond).C
			tt.c.Broadcast()
			cancel()
			select {
			case <-time.NewTimer(500 * time.Millisecond).C:
				t.Error("Goroutine timeout error")
			case <-wrapWait(tt.wg):
				if _, ok := tt.s.GaugeMetrics["Alloc"]; !ok {
					t.Error("Failed to collect data")
				}
			}
		})
	}
}

func Test_pollPSUtilMetrics(t *testing.T) {
	tests := []struct {
		ctx  context.Context
		c    *sync.Cond
		s    *storage.MemStorage
		wg   *sync.WaitGroup
		name string
	}{
		{
			name: "context exit, data collected",
			c:    sync.NewCond(&sync.Mutex{}),
			s:    storage.NewMemStorage(),
			wg:   new(sync.WaitGroup),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cancel context.CancelFunc
			tt.wg.Add(1)
			tt.ctx, cancel = context.WithCancel(context.Background())
			go pollPSUtilMetrics(tt.ctx, tt.c, tt.s, tt.wg)
			<-time.NewTimer(100 * time.Millisecond).C
			tt.c.Broadcast()
			cancel()
			select {
			case <-time.NewTimer(500 * time.Millisecond).C:
				t.Error("Goroutine timeout error")
			case <-wrapWait(tt.wg):
				if _, ok := tt.s.GaugeMetrics["TotalMemory"]; !ok {
					t.Error("Failed to collect data")
				}
			}
		})
	}
}
