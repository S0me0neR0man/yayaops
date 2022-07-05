package client

import (
	"context"
	"fmt"
	"github.com/S0me0neR0man/yayaops/internal/common"
	"log"
	"math/rand"
	"net/http"
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
	sync.RWMutex
	gauges   common.GaugesData
	counters common.CountersData
	counter  common.Counter
}

var metrics = &metricsEngine{}

func GetEngine() *metricsEngine {
	return metrics
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
	g := m.getGauges()
	c := m.getCounters()
	m.Lock()
	m.gauges = g // todo: correct?
	m.counters = c
	m.Unlock()
}

func (m *metricsEngine) getGauges() (res common.GaugesData) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	for i, metric := range common.GaugeMetrics {
		switch metric {
		case "Alloc":
			res[i] = common.Gauge(ms.Alloc)
		case "BuckHashSys":
			res[i] = common.Gauge(ms.BuckHashSys)
		case "Frees":
			res[i] = common.Gauge(ms.Frees)
		case "GCCPUFraction":
			res[i] = common.Gauge(ms.GCCPUFraction)
		case "GCSys":
			res[i] = common.Gauge(ms.GCSys)
		case "HeapAlloc":
			res[i] = common.Gauge(ms.HeapAlloc)
		case "HeapIdle":
			res[i] = common.Gauge(ms.HeapIdle)
		case "HeapInuse":
			res[i] = common.Gauge(ms.HeapInuse)
		case "HeapObjects":
			res[i] = common.Gauge(ms.HeapObjects)
		case "HeapReleased":
			res[i] = common.Gauge(ms.HeapReleased)
		case "HeapSys":
			res[i] = common.Gauge(ms.HeapSys)
		case "LastGC":
			res[i] = common.Gauge(ms.LastGC)
		case "Lookups":
			res[i] = common.Gauge(ms.Lookups)
		case "MCacheInuse":
			res[i] = common.Gauge(ms.MCacheInuse)
		case "MCacheSys":
			res[i] = common.Gauge(ms.MCacheSys)
		case "MSpanInuse":
			res[i] = common.Gauge(ms.MSpanInuse)
		case "MSpanSys":
			res[i] = common.Gauge(ms.MSpanSys)
		case "Mallocs":
			res[i] = common.Gauge(ms.Mallocs)
		case "NextGC":
			res[i] = common.Gauge(ms.NextGC)
		case "NumForcedGC":
			res[i] = common.Gauge(ms.NumForcedGC)
		case "NumGC":
			res[i] = common.Gauge(ms.NumGC)
		case "OtherSys":
			res[i] = common.Gauge(ms.OtherSys)
		case "PauseTotalNs":
			res[i] = common.Gauge(ms.PauseTotalNs)
		case "StackInuse":
			res[i] = common.Gauge(ms.StackInuse)
		case "StackSys":
			res[i] = common.Gauge(ms.StackSys)
		case "Sys":
			res[i] = common.Gauge(ms.Sys)
		case "TotalAlloc":
			res[i] = common.Gauge(ms.TotalAlloc)
		case "RandomValue":
			res[i] = common.Gauge(rand.Float64())
		}
	}
	return res
}

func (m *metricsEngine) getCounters() (res common.CountersData) {
	res[0] = m.counter
	m.counter++
	return res
}

func (m *metricsEngine) sendReport() {
	m.RLock()
	g := m.gauges
	c := m.counters
	m.RUnlock()
	clnt := http.Client{}
	for i, metric := range common.GaugeMetrics {
		url := fmt.Sprintf("http://%s/update/gauge/%s/%f", addr, metric, g[i])
		log.Println(url)
		r, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			continue
		}
		response, err := clnt.Do(r)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(response)
	}
	url := fmt.Sprintf("http://%s/update/counter/PollCount/%d", addr, c[0])
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}
	response, err := clnt.Do(r)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(response)
}
