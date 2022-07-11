package client

import (
	"github.com/S0me0neR0man/yayaops/internal/common"
	"testing"
)

func Test_metricType(t *testing.T) {
	// gauge
	t.Run("#1 gauge", func(t *testing.T) {
		if got := typeOfValue(common.Gauge(0)); got != "gauge" {
			t.Errorf("typeOfValue() = %v, want %v", got, "gauge")
		}
	})
	// counter
	t.Run("#2 counter", func(t *testing.T) {
		if got := typeOfValue(common.Counter(0)); got != "counter" {
			t.Errorf("typeOfValue() = %v, want %v", got, "counter")
		}
	})
}
