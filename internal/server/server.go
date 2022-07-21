package server

import (
	"fmt"
	"github.com/S0me0neR0man/yayaops/internal/common"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

const (
	addr = "127.0.0.1:8080"

	TypeGauge   = "gauge"
	TypeCounter = "counter"

	OperUpdateMetric = "update"
	OperGetMetric    = "value"

	MuxOper  = "oper"
	MuxMType = "type"
	MuxMName = "metric"
	MuxValue = "value"
)

type Server struct {
	storage *common.Storage
}

func New() *Server {
	s := Server{}
	s.storage = common.NewStorage()
	return &s
}

// Start set handlers and start listening
func (s *Server) Start() error {
	router := mux.NewRouter()
	s.setHandlers(router)

	return http.ListenAndServe(addr, router)
}

func (s *Server) setHandlers(router *mux.Router) {
	router.Use(s.logging)

	router.HandleFunc("/{oper}/{type}/{metric}/{value}",
		s.metricsPostHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/text", "Content-Type", "text/plain")
	router.HandleFunc("/{oper}/{type}/{metric}",
		s.metricsGetHandler).
		Methods(http.MethodGet).
		Headers("Content-Type", "application/text", "Content-Type", "text/plain")

	router.HandleFunc("/{oper}/{type}/{metric}/{value}", s.notAcceptableHandler)
	router.HandleFunc("/{oper}/{type}/{metric}", s.notAcceptableHandler)
}

// logging middleware
func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.Header, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// notAcceptableHandler the handler of incorrect requests
func (s *Server) notAcceptableHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Allowed", http.StatusNotAcceptable)
}

// metricsPostHandler
func (s *Server) metricsPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if cmd, status := newCommand(vars); status != http.StatusOK {
		w.WriteHeader(status)
		return
	} else {
		// save
		switch cmd.MType {
		case TypeGauge:
			s.storage.Set(cmd.ID, *cmd.Value)
		case TypeCounter:
			if old, ok := s.storage.Get(cmd.ID); ok {
				s.storage.Set(cmd.ID, old, *cmd.Delta)
			} else {
				s.storage.Set(cmd.ID, *cmd.Delta)
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

// metricsGetHandler
func (s *Server) metricsGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if cmd, status := newCommand(vars); status != http.StatusOK {
		w.WriteHeader(status)
		return
	} else {
		if v, ok := s.storage.Get(cmd.ID); ok {
			if _, err := w.Write([]byte(fmt.Sprintf("%v", v))); err != nil {
				log.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

func newCommand(vars map[string]string) (*common.Command, int) {
	c := &common.Command{CType: common.CTUnknown, Metrics: common.Metrics{Delta: new(int64), Value: new(float64)}}

	for key, val := range vars {
		switch key {
		case MuxOper: // operation
			switch val {
			case OperUpdateMetric:
				c.CType = common.CTUpdate
			case OperGetMetric:
				c.CType = common.CTGet
			default:
				return nil, http.StatusNotFound
			}
		case MuxMType: // metric type
			switch val {
			case TypeGauge:
				c.MType = TypeGauge
			case TypeCounter:
				c.MType = TypeCounter
			default:
				return nil, http.StatusNotImplemented
			}
		case MuxMName: // metric name
			c.ID = val
		}
	}
	if c.CType == common.CTUnknown || c.MType == "" || c.ID == "" {
		return nil, http.StatusBadRequest
	}

	if c.CType == common.CTUpdate {
		switch c.MType {
		case TypeGauge:
			if v, err := strconv.ParseFloat(vars[MuxValue], 64); err == nil {
				*c.Value = v
			} else {
				return nil, http.StatusBadRequest
			}
		case TypeCounter:
			if v, err := strconv.Atoi(vars[MuxValue]); err == nil {
				*c.Delta = int64(v)
			} else {
				return nil, http.StatusBadRequest
			}
		}
	}
	return c, http.StatusOK
}
