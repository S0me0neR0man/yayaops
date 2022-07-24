package common

import (
	"errors"
	"reflect"
	"strconv"
)

const (
	CTUnknown = iota
	CTUpdate
	CTValue

	MTypeGauge   = "gauge"
	MTypeCounter = "counter"
)

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
	return append(RuntimeMNames, CustomMNames...)
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// SetStrValue MType must be filled before the call
func (m *Metrics) SetStrValue(value string) error {
	switch m.MType {
	case MTypeGauge:
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			if m.Value == nil {
				m.Value = new(float64)
			}
			*m.Value = v
		} else {
			return err
		}
	case MTypeCounter:
		if v, err := strconv.Atoi(value); err == nil {
			if m.Delta == nil {
				m.Delta = new(int64)
			}
			*m.Delta = int64(v)
		} else {
			return err
		}
	}
	return nil
}

// SetAnyValue MType must be filled before the call
// value must be int64 or float64
func (m *Metrics) SetAnyValue(value any) error {
	v := reflect.ValueOf(value)
	if m.MType == MTypeGauge {
		if v.CanFloat() {
			if m.Value == nil {
				m.Value = new(float64)
			}
			*m.Value = v.Float()
			return nil
		}
	} else {
		if v.CanInt() {
			if m.Delta == nil {
				m.Delta = new(int64)
			}
			*m.Delta = v.Int()
			return nil
		}
	}
	return errors.New("SetAnyValue: wrong  type")
}

type Command struct {
	Metrics
	CType    int  `json:"-"`
	JSONResp bool `json:"-"`
}
