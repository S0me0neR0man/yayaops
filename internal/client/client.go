package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/S0me0neR0man/yayaops/internal/common"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type config struct {
	addr           string
	pollInterval   time.Duration
	reportInterval time.Duration
}

var cfg config

func init() {
	if cfg.addr = os.Getenv("ADDRESS"); cfg.addr == "" {
		cfg.addr = "127.0.0.1:8080"
	}
	cfg.pollInterval = 2 * time.Second
	if s := os.Getenv("POOL_INTERVAL"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			cfg.pollInterval = time.Duration(v) * time.Second
		}
	}
	cfg.reportInterval = 10 * time.Second
	if s := os.Getenv("REPORT_INTERVAL"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			cfg.reportInterval = time.Duration(v) * time.Second
		}
	}

	log.Printf("Client init %+v", cfg)
}

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
	ticker := time.NewTicker(cfg.pollInterval)
	// start reportJob goroutine
	ctxReport, cancelReport := context.WithCancel(ctx)
	m.wg.Add(1)
	go m.reportJob(ctxReport)
	for {
		select {
		case <-ticker.C:
			m.pollMetrics()
		case <-ctx.Done():
			ticker.Stop()
			cancelReport()
			m.wg.Done()
			return
		}
	}
}

// reportJob goroutine for send report
func (m *metricsEngine) reportJob(ctx context.Context) {
	ticker := time.NewTicker(cfg.reportInterval)
	for {
		select {
		case <-ticker.C:
			m.sendReport()
		case <-ctx.Done():
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
		v := reflect.ValueOf(ms).FieldByName(name)
		switch v.Kind() {
		case reflect.Int64:
			_ = m.storage.Set(name, v.Int())
		case reflect.Uint64, reflect.Uint32:
			_ = m.storage.Set(name, int64(v.Uint()))
		case reflect.Float64:
			_ = m.storage.Set(name, v.Float())
		default:
			log.Println("pollMetrics", v.Kind())
		}
	}
	// custom
	_ = m.storage.Set("RandomValue", rand.Float64())
	_ = m.storage.Set("PollCount", m.pollCount)
	m.pollCount++
}

func (m *metricsEngine) sendReport() {
	c := resty.New()
	for _, name := range m.storage.GetNames() {
		if val, ok := m.storage.Get(name); ok {
			m := common.Metrics{}
			//m.MType = typeOfMetric(val)
			if name == "PollCount" {
				m.MType = common.MTypeCounter
			} else {
				m.MType = common.MTypeGauge
			}
			m.ID = name
			if err := m.SetAnyValue(val); err != nil {
				log.Println(err)
			}
			b, _ := json.Marshal(m)
			log.Println(string(b))
			url := fmt.Sprintf("http://%s/update/", cfg.addr)
			resp, err := c.R().SetHeader("Content-Type", "application/json").SetBody(b).Post(url)
			if err != nil {
				log.Println(resp, err)
			}
		}
	}
}

func typeOfMetric(val any) string {
	switch v := reflect.ValueOf(val); v.Kind() {
	case reflect.Float64:
		return common.MTypeGauge
	case reflect.Int64:
		return common.MTypeCounter
	default:
		log.Fatal("client.typeOfMetric() unknown metric type:", v.Kind().String())
	}
	return "unknown"
}
