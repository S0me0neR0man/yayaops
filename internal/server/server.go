package server

import (
	"encoding/json"
	"fmt"
	"github.com/S0me0neR0man/yayaops/internal/common"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	addr = "127.0.0.1:8080"

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

// setHandlers configure gorilla/mux router
func (s *Server) setHandlers(router *mux.Router) {
	router.Use(s.logging)

	router.HandleFunc("/update/", s.updateJSONHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")
	router.HandleFunc("/value/", s.valueJSONHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	router.HandleFunc("/{oper}/{type}/{metric}/{value}", s.postHandler).
		Methods(http.MethodPost)
	router.HandleFunc("/{oper}/{type}/{metric}", s.getHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/{oper}/{type}/{metric}/{value}", s.notAcceptableHandler)
	router.HandleFunc("/{oper}/{type}/{metric}", s.notAcceptableHandler)
}

// logging middleware
func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.Header, r.RequestURI, r.Body)
		next.ServeHTTP(w, r)
	})
}

// notAcceptableHandler the handler of incorrect requests
func (s *Server) notAcceptableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
}

// postHandler http.POST without 'Content-Type'
func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if cmd, status := commandFromURL(vars); status == http.StatusOK {
		s.executeCommand(cmd, w)
	} else {
		w.WriteHeader(status)
	}
}

// getHandler http.GET without 'Content-Type'
func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if cmd, status := commandFromURL(vars); status == http.StatusOK {
		s.executeCommand(cmd, w)
	} else {
		w.WriteHeader(status)
	}
}

// updateJSONHandler POST update/
func (s *Server) updateJSONHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var b []byte

	if b, err = ioutil.ReadAll(r.Body); err == nil {
		log.Println(string(b))
		m := common.Metrics{}
		if err = json.Unmarshal(b, &m); err == nil {
			cmd := common.Command{Metrics: m, CType: common.CTUpdate, JSONResp: true}
			s.executeCommand(&cmd, w)
			return
		}
	}
	log.Println(err)
}

// valueJSONHandler POST value/
func (s *Server) valueJSONHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var b []byte

	if b, err = ioutil.ReadAll(r.Body); err == nil {
		log.Println(string(b))
		m := common.Metrics{}
		if err = json.Unmarshal(b, &m); err == nil {
			cmd := common.Command{Metrics: m, CType: common.CTValue, JSONResp: true}
			s.executeCommand(&cmd, w)
			return
		}
	}
	log.Println(err)
}

func (s *Server) executeCommand(cmd *common.Command, w http.ResponseWriter) {
	if cmd.JSONResp {
		w.Header().Set("Content-Type", "application/json")
	}
	switch cmd.CType {
	case common.CTUpdate:
		switch cmd.MType {
		case common.MTypeGauge:
			s.storage.Set(cmd.ID, *cmd.Value)
		case common.MTypeCounter:
			if old, ok := s.storage.Get(cmd.ID); ok {
				s.storage.Set(cmd.ID, old, *cmd.Delta)
			} else {
				s.storage.Set(cmd.ID, *cmd.Delta)
			}
		default:
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		w.WriteHeader(http.StatusOK)
	case common.CTValue:
		if v, ok := s.storage.Get(cmd.ID); ok {
			var b []byte
			var err error
			if cmd.JSONResp {
				cmd.SetAnyValue(v)
				b, err = json.Marshal(cmd)
			} else {
				b = []byte(fmt.Sprintf("%v", v))
			}
			if err == nil {
				_, err = w.Write(b)
			}
			if err != nil {
				log.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}

// commandFromURL sprint 1 compatibility with sprint 2
func commandFromURL(vars map[string]string) (*common.Command, int) {
	c := &common.Command{CType: common.CTUnknown}

	for key, val := range vars {
		switch key {
		case MuxOper:
			switch val {
			case OperUpdateMetric:
				c.CType = common.CTUpdate
			case OperGetMetric:
				c.CType = common.CTValue
			default:
				return nil, http.StatusNotFound
			}
		case MuxMType:
			switch val {
			case common.MTypeGauge:
				c.MType = common.MTypeGauge
			case common.MTypeCounter:
				c.MType = common.MTypeCounter
			default:
				return nil, http.StatusNotImplemented
			}
		case MuxMName:
			c.ID = val
		}
	}
	if c.CType == common.CTUnknown || c.MType == "" || c.ID == "" {
		return nil, http.StatusBadRequest
	}
	if c.CType == common.CTUpdate {
		if err := c.SetStrValue(vars[MuxValue]); err != nil {
			return nil, http.StatusBadRequest
		}
	}
	return c, http.StatusOK
}
