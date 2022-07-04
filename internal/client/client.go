package client

import (
	"sync"
	"time"
	"yayaops/common"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

type metricsEngine struct {
	sync.RWMutex
	gauges       map[string]common.Gauge
	counters     map[string]common.Counter
	pollTicker   *time.Ticker
	reportTicker *time.Ticker
	pollStop     chan interface{}
	reportStop   chan interface{}
}

var metrics = &metricsEngine{}

func GetEngine() *metricsEngine {
	return metrics
}

func (m *metricsEngine) Start() error {
	go m.poll()
	go m.report()
	m.pollTicker = time.NewTicker(pollInterval)
	m.reportTicker = time.NewTicker(reportInterval)
	return nil
}

func (m *metricsEngine) Stop() {
	m.reportStop <- 1
	m.pollStop <- 1
}

func (m *metricsEngine) poll() {
	for {
		select {
		case <-m.pollTicker.C:
			m.Lock()
			m.Unlock()
		case <-m.pollStop:
			m.reportTicker.Stop()
			return
		}
	}
}

func (m *metricsEngine) report() {
	for {
		select {
		case <-m.reportTicker.C:
			m.RLock()
			m.RUnlock()
		case <-m.reportStop:
			m.reportTicker.Stop()
			return
		}
	}
}
