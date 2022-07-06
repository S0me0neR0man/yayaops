package client

import (
	"context"
	"fmt"
	"github.com/S0me0neR0man/yayaops/internal/common"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	addr           = "127.0.0.1:8080"
)

type metricsEngine struct {
	gauges    *common.Storage[common.Gauge]
	counters  *common.Storage[common.Counter]
	pollCount int64
}

func GetEngine() *metricsEngine {
	e := metricsEngine{}
	e.gauges = common.NewStorage[common.Gauge]()
	e.counters = common.NewStorage[common.Counter]()
	return &e
}

// Start engine
func (m *metricsEngine) Start(ctx context.Context) {
	go m.poll(ctx)
}

// poll metrics collection goroutine
func (m *metricsEngine) poll(ctx context.Context) {
	// get metrics
	m.fillMetrics()
	ticker := time.NewTicker(pollInterval)
	// start report goroutine
	ctxReport, cancelReport := context.WithCancel(ctx)
	go m.report(ctxReport)
	for {
		select {
		case <-ticker.C:
			m.fillMetrics()
			log.Println("### poll", m.gauges, m.counters)
		case <-ctx.Done():
			// todo: корректное завершение, сюда не доходит
			log.Println("exit POLL")
			ticker.Stop()
			cancelReport()
			return
		}
	}
}

// report - metrics sending goroutine
func (m *metricsEngine) report(ctx context.Context) {
	ticker := time.NewTicker(reportInterval)
	for {
		select {
		case <-ticker.C:
			log.Println("@@@ report")
			m.sendReport()
		case <-ctx.Done():
			// todo: корректное завершение, сюда не доходит
			log.Println("exit report")
			ticker.Stop()
			return
		}
	}
}

func (m *metricsEngine) fillMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	for _, name := range common.RuntimeMNames {
		switch v := reflect.ValueOf(ms).FieldByName(name); v.Kind() {
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			m.gauges.Set(name, common.Gauge(float64(v.Uint())))
		case reflect.Float64:
			m.gauges.Set(name, common.Gauge(v.Float()))
		}
	}
	m.gauges.Set("RandomValue", common.Gauge(rand.Float64()))
	m.counters.Set("PollCount", common.Counter(m.pollCount))
	m.pollCount++
}

func (m *metricsEngine) sendReport() {
	sendStorage[common.Gauge](m.gauges)
	sendStorage[common.Counter](m.counters)
}

func sendStorage[T common.Metric](storage common.Getter[T]) {
	c := http.Client{}
	for _, name := range storage.GetNames() {
		if val, ok := storage.Get(name); ok {
			url := fmt.Sprintf("http://%s/update/gauge/%s/%s", addr, name, val.String())
			r, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println(url, err)
				continue
			}
			if response, err := c.Do(r); err != nil {
				log.Println(url, response, err)
				continue
			}
		} else {
			log.Println("sendGauges error: metric", name, "not found")
		}
	}
}
