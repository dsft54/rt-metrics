package scheduller

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dsft54/rt-metrics/config/agent/settings"
)

func TestNewScheduller(t *testing.T) {
	tests := []struct {
		name string
		cfg  *settings.Config
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
	tsch := &Scheduller{
		Rc: sync.NewCond(&sync.Mutex{}),
		Pc: sync.NewCond(&sync.Mutex{}),
	}
	wrapChan := wrapSync(tsch)
	select {
	case <-wrapChan:
		// Ok
	case <-time.NewTimer(500 * time.Millisecond).C:
		t.Fail()
	}
}

func wrapSync(sch *Scheduller) <-chan struct{} {
	go func() {
		<-time.NewTimer(200 * time.Millisecond).C
		fmt.Println("Should be released")
		sch.ExitRelease()
	}()
	out := make(chan struct{})
	sch.Rc.L.Lock()
	defer sch.Rc.L.Unlock()
	sch.Rc.Wait()
	fmt.Println("-------------------------------------------------Should be released")
	out <- struct{}{}
	return out
}
