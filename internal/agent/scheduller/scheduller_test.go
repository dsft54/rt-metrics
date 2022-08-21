package scheduller

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dsft54/rt-metrics/config/agent/settings"
)

func TestNewScheduller(t *testing.T) {
	tests := []struct {
		cfg  *settings.Config
		name string
		want bool
	}{
		{
			name: "Basic test",
			cfg: &settings.Config{
				PollInterval:   1,
				ReportInterval: 1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScheduller(tt.cfg); !reflect.DeepEqual(got.Update, tt.want) {
				t.Errorf("NewScheduller() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduller_Start(t *testing.T) {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	tsch := NewScheduller(&settings.Config{
		PollInterval:   1,
		ReportInterval: 1,
	})
	go tsch.Start(ctx, wg)
	cancel()
	select {
	case <-wrapWait(wg):
		// Ok
	case <-time.NewTimer(500 * time.Millisecond).C:
		t.Fail()
	}
}

func wrapWait(wg *sync.WaitGroup) <-chan struct{} {
	out := make(chan struct{})
	go func() {
		wg.Wait()
		out <- struct{}{}
	}()
	return out
}

func TestScheduller_ExitRelease(t *testing.T) {
	tests := []struct {
		sch  *Scheduller
		name string
	}{
		{
			name: "release test",
			sch: NewScheduller(&settings.Config{
				ReportInterval: 1,
				PollInterval:   1,
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				<-time.NewTimer(100 * time.Millisecond).C
				tt.sch.ExitRelease()

			}()
			go func() {
				<-time.NewTimer(200 * time.Millisecond).C
				tt.sch.ExitRelease()

			}()
			go func() {
				<-time.NewTimer(1000 * time.Millisecond).C
				tt.sch.Pc.Signal()
				tt.sch.Rc.Signal()
				t.Error("Timeout")
			}()
			tt.sch.Pc.L.Lock()
			tt.sch.Pc.Wait()
			tt.sch.Pc.L.Unlock()
			tt.sch.Rc.L.Lock()
			tt.sch.Rc.Wait()
			tt.sch.Rc.L.Unlock()
			if tt.sch.Update != false {
				t.Errorf("Scheduller not working")
			}
		})
	}
}
