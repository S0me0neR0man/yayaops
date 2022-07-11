package server

import (
	"github.com/S0me0neR0man/yayaops/internal/common"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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

func (s *Server) Start() error {
	router := mux.NewRouter()

	// middleware
	router.Use(s.logging)

	router.HandleFunc("/{oper}/{type}/{metric}/{value}", s.metricsPostHandler).Methods(http.MethodPost)
	router.HandleFunc("/{oper}/{type}/{metric}", s.metricsGetHandler).Methods(http.MethodGet)

	router.HandleFunc("/{oper}/{type}/{metric}/{value}", s.notAcceptableHandler)
	router.HandleFunc("/{oper}/{type}/{metric}", s.notAcceptableHandler)

	return http.ListenAndServe(addr, router)
}

// logging middleware
func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// notAllowedHandler the handler of incorrect requests
func (s *Server) notAcceptableHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Allowed", http.StatusNotAcceptable)
}

// metricsPostHandler
func (s *Server) metricsPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// checks
	if status := checkURI(vars); status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	// Ok, just do it
	if vars["value"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch vars["type"] {
	case "gauge":
		if v, err := common.Gauge(0).From(vars["value"]); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			s.gauges.Set(vars["metric"], v.(common.Gauge))
		}
	case "counter":
		if v, err := common.Counter(0).From(vars["value"]); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			if old, ok := s.counters.Get(vars["metric"]); ok {
				s.counters.Set(vars["metric"], old+v.(common.Counter))
			} else {
				s.counters.Set(vars["metric"], v.(common.Counter))
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// metricsPostHandler
func (s *Server) metricsGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// checks
	if status := checkURI(vars); status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	// Ok, just do it
	switch vars["type"] {
	case "gauge":
		if v, ok := s.gauges.Get(vars["metric"]); ok {
			w.Write([]byte(v.String()))
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	case "counter":
		if v, ok := s.counters.Get(vars["metric"]); ok {
			w.Write([]byte(v.String()))
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func checkURI(vars map[string]string) int {
	for key, val := range vars {
		switch key {
		case "oper": // operation
			switch val {
			case "update":
			case "value":
			default:
				return http.StatusNotFound
			}
		case "type": // metric type
			switch val {
			case "gauge":
			case "counter":
			default:
				return http.StatusNotImplemented
			}
		case "metric": // metric name
			if val == "" {
				return http.StatusNotFound
			}
		}
	}
	return http.StatusOK
}
