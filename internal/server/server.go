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

	for i, p := range path {
		switch i {
		case 0:
		case 1:
			if p != "update" {
				return http.StatusNotFound
			}
		case 2:
			switch p {
			case "gauge", "counter":
			default:
				return http.StatusNotFound
			}
		case 3:
			fflag := false
			for _, name := range common.RuntimeMNames {
				if p == name {
					fflag = true
					break
				}
			}
			if !fflag {
				return http.StatusNotFound
			}
		case 4:
			if p != "" {
				res = http.StatusOK
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
