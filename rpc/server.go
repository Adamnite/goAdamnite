package rpc

import (
	"sync/atomic"

	"github.com/adamnite/go-adamnite/log15"
)

// RPC server
type Server struct {
	services adamniteServiceRegistry
	run      int32
}

func NewAdamniteRPCServer() *Server {
	server := &Server{run: 1}

	return server
}

func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log15.Debug("RPC server shutting down")
	}
}

func (s *Server) RegisterName(name string, receiver interface{}) error {
	return s.services.registerName(name, receiver)
}
