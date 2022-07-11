package common

import (
	"fmt"
	"strconv"
	"sync"
)

type ValueFromString interface {
	From(string) (any, error)
}

// Gauge metrics
type Gauge float64

func (g Gauge) String() string {
	return fmt.Sprintf("%f", g)
}

func (g Gauge) From(s string) (any, error) {
	if v, err := strconv.ParseFloat(s, 64); err != nil {
		return nil, err
	} else {
		return Gauge(v), nil
	}
}

// Counter metrics
type Counter int64

func (c Counter) String() string {
	return fmt.Sprintf("%d", c)
}

func (c Counter) From(s string) (any, error) {
	if v, err := strconv.Atoi(s); err != nil {
		return nil, err
	} else {
		return Counter(v), nil
	}
}

// Metric generic metric
type Metric interface {
	Gauge | Counter
	fmt.Stringer
}

var RuntimeMNames = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

var CustomMNames = []string{
	"RandomValue",
	"PollCount",
}

func ValidMetricsName() []string {
	return append(RuntimeMNames, CustomMNames[0], CustomMNames[1])
}

// Getter interface for get Metric
type Getter[T Metric] interface {
	Get(string) (T, bool)
	GetNames() []string
}

// Setter interface for set Metric
type Setter[T Metric] interface {
	Set(string, T)
}

// Storage the thread-safe storage
type Storage[T Metric] struct {
	sync.RWMutex
	data map[string]T
}

// New the constructor
func New[T Metric]() *Storage[T] {
	s := Storage[T]{}
	s.data = make(map[string]T)
	return &s
}

// Set implementation the Setter
func (s *Storage[T]) Set(key string, value T) {
	s.Lock()
	s.data[key] = value
	s.Unlock()
}

// Get implementation the Getter
func (s *Storage[T]) Get(name string) (T, bool) {
	s.RLock()
	v, ok := s.data[name]
	s.RUnlock()
	return v, ok
}

func (s *Storage[T]) GetNames() []string {
	names := make([]string, len(s.data))
	i := 0
	s.RLock()
	for k := range s.data {
		names[i] = k
		i++
	}
	s.RUnlock()
	return names
}
