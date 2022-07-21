package client

import (
	"context"
	"fmt"
	"github.com/S0me0neR0man/yayaops/internal/common"
	"github.com/S0me0neR0man/yayaops/internal/server"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	addr           = "127.0.0.1:8080"
)

type metricsEngine struct {
	storage   *common.Storage
	pollCount int64
	wg        sync.WaitGroup
}

func New() *metricsEngine {
	e := metricsEngine{}
	e.storage = common.NewStorage()
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
			log.Println("### pollJob", m.storage)
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
	// runtime metrics
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	for _, name := range common.RuntimeMNames {
		m.storage.Set(name, reflect.ValueOf(ms).FieldByName(name))
	}
	// custom
	m.storage.Set("RandomValue", rand.Float64())
	m.storage.Set("PollCount", m.pollCount)
	m.pollCount++
}

func (m *metricsEngine) sendReport() {
	c := http.Client{}
	for _, name := range m.storage.GetNames() {
		if val, ok := m.storage.Get(name); ok {
			mt := typeOfMetric(val)
			url := fmt.Sprintf("http://%s/update/%s/%s/%v", addr, mt, name, val)
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

func typeOfMetric(val any) string {
	switch v := reflect.ValueOf(val); v.Kind() {
	case reflect.Float64:
		return server.TypeGauge
	case reflect.Int64, reflect.Uint32:
		return server.TypeCounter
	default:
		log.Println("client.typeOfMetric() illegal type", v.Kind().String())
	}
	return "unknown"
}
