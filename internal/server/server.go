package server

import (
	"github.com/S0me0neR0man/yayaops/internal/common"
	"log"
	"net/http"
	"strings"
)

const (
	addr = "127.0.0.1:8080"
)

type Server struct {
	gauges   *common.Storage[common.Gauge]
	counters *common.Storage[common.Counter]
}

func New() *Server {
	s := Server{}
	s.gauges = common.New[common.Gauge]()
	s.counters = common.New[common.Counter]()
	return &s
}

func (s *Server) Start() {
	http.HandleFunc("/", oneForAllHandler)
	server := &http.Server{Addr: addr}
	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

func parseURI(uri string) int {
	res := http.StatusNotAcceptable
	path := strings.Split(uri, "/")
	var metric any

	for i, p := range path {
		switch i {
		case 1: // operation
			if p != "update" {
				return http.StatusNotFound
			}
		case 2: // type
			switch p {
			case "gauge":
				metric = common.Gauge(0)
			case "counter":
				metric = common.Counter(0)
			default:
				return http.StatusNotFound
			}
		case 3: // metric name
			if p == "" {
				return http.StatusBadRequest
			}
		case 4: // value
			if p == "" {
				res = http.StatusBadRequest
			}
			switch v, ok := metric.(common.Gauge); ok {
			case true: // Gauge
				if _, err := v.FromString(p); err != nil {
					res = http.StatusBadRequest
				} else {
					res = http.StatusOK
				}
			case false: // Counter
				if _, err := metric.(common.Counter).FromString(p); err != nil {
					res = http.StatusBadRequest
				} else {
					res = http.StatusOK
				}
			}
			if v, ok := metric.(common.Counter); ok {
				if _, err := v.FromString(p); err == nil {
					return http.StatusOK
				} else {
					return http.StatusBadRequest
				}
			}
		}
	}
	return res
}

func oneForAllHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	switch r.Method {
	case "POST":
		w.WriteHeader(parseURI(r.RequestURI))
		_, err := w.Write([]byte(r.RequestURI))
		if err != nil {
			log.Println(err)
			return
		}
	default:
		w.WriteHeader(http.StatusNotAcceptable)
	}
}
