package server

import (
	"log"
	"net/http"
	"time"
)

type Server struct {
	addr string
}

func New(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Println("server running on", s.addr)
	return srv.ListenAndServe()
}