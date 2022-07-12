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
	"strings"
	"sync"
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
	wg        sync.WaitGroup
}

func New() *metricsEngine {
	e := metricsEngine{}
	e.gauges = common.New[common.Gauge]()
	e.counters = common.New[common.Counter]()
	return &e
}

// Start engine
func (m *metricsEngine) Start(ctx context.Context) *metricsEngine {
	m.wg.Add(1)
	go m.pollJob(ctx)
	return m
}

// WaitShutdown wait for a stop goroutines
func (m *metricsEngine) WaitShutdown() {
	m.wg.Wait()
}

// pollJob goroutine for collect metrics
func (m *metricsEngine) pollJob(ctx context.Context) {
	// get metrics
	m.pollMetrics()
	ticker := time.NewTicker(pollInterval)
	// start reportJob goroutine
	ctxReport, cancelReport := context.WithCancel(ctx)
	m.wg.Add(1)
	go m.reportJob(ctxReport)
	for {
		select {
		case <-ticker.C:
			m.pollMetrics()
			log.Println("### pollJob", m.gauges, m.counters)
		case <-ctx.Done():
			log.Println("--- exit POLL")
			ticker.Stop()
			cancelReport()
			m.wg.Done()
			return
		}
	}
}

// reportJob goroutine for send report
func (m *metricsEngine) reportJob(ctx context.Context) {
	ticker := time.NewTicker(reportInterval)
	for {
		select {
		case <-ticker.C:
			log.Println("@@@ reportJob")
			m.sendReport()
		case <-ctx.Done():
			log.Println("--- exit reportJob")
			ticker.Stop()
			m.wg.Done()
			return
		}
	}
}

func (m *metricsEngine) pollMetrics() {
	// runtime Gauges
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
	// custom Gauges
	m.gauges.Set("RandomValue", common.Gauge(rand.Float64()))
	// custom Counters
	m.counters.Set("PollCount", common.Counter(m.pollCount))
	m.pollCount++
}

func (m *metricsEngine) sendReport() {
	sendFromStorage[common.Gauge](m.gauges)
	sendFromStorage[common.Counter](m.counters)
}

func sendFromStorage[T common.Metric](storage common.Getter[T]) {
	//	mt := metricType[T]()
	c := http.Client{}
	for _, name := range storage.GetNames() {
		if val, ok := storage.Get(name); ok {
			mt := typeOfValue(val)
			url := fmt.Sprintf("http://%s/update/%s/%s/%s", addr, mt, name, val.String())
			r, err := http.NewRequest("POST", url, nil)
			if err != nil {
				log.Println(url, err)
				continue
			}
			if resp, err := c.Do(r); err != nil {
				log.Println(url, resp, err)
			}
		} else {
			log.Println("sendGauges error: metric", name, "not found")
		}
	}
}

func typeOfValue(val any) string {
	ss := strings.Split(reflect.ValueOf(val).Type().String(), ".")
	return strings.ToLower(ss[len(ss)-1])
}
