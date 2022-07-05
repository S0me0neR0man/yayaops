package server

import (
	"log"
	"net/http"
)

const (
	addr = "127.0.0.1:8080"
)

type Server struct {
}

var s = &Server{}

func GetServer() *Server {
	return s
}

func (s *Server) Start() {
	http.HandleFunc("/", s.allHandler)
	server := &http.Server{Addr: addr}
	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) allHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	_, err := w.Write([]byte(r.RequestURI))
	if err != nil {
		log.Println(err)
		return
	}
}
