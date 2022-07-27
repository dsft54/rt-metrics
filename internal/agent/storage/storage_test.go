// Модуль storage определяет структуры их методы, предназначенные для описания хранилища текущего значения метрик,

// из которого они будут отправлены на сервер.

package storage

import "fmt"

func ExampleMemStorage_ConvertToURLParams() {
	ms := &MemStorage{
		GaugeMetrics:   map[string]gauge{"Alloc": 3.14},
		CounterMetrics: map[string]counter{},
	}
	out1 := ms.ConvertToURLParams()
	fmt.Println(out1)

	ms = &MemStorage{
		GaugeMetrics:   map[string]gauge{"Alloc": 3.14, "Heap": 6.28},
		CounterMetrics: map[string]counter{"Counter": 1},
	}
	out2 := ms.ConvertToURLParams()
	fmt.Println(out2)

	// Output:
	// [/gauge/Alloc/3.14]
	// [/gauge/Alloc/3.14 /gauge/Heap/6.28 /counter/Counter/1]
}
