package rpc

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/adamnite/go-adamnite/log15"
	mapset "github.com/deckarep/golang-set"
)

type CodecOption int

// RPC server
type Server struct {
	services adamniteServiceRegistry
	run      int32
	idgen    func() ID
	codecs   mapset.Set
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

func (s *Server) serveSingleRequest(ctx context.Context, codec ServerCodec) {
	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	h := newHandler(ctx, codec, s.idgen, &s.services)
	h.allowSubscribe = false
	defer h.close(io.EOF, nil)

	reqs, batch, err := codec.readBatch()
	if err != nil {
		if err != io.EOF {
			codec.writeJSON(ctx, errorMessage(&invalidMessageError{"parse error"}))
		}
		return
	}
	if batch {
		h.handleBatch(reqs)
	} else {
		h.handleMsg(reqs[0])
	}
}

func (s *Server) ServeCodec(codec ServerCodec, options CodecOption) {
	defer codec.close()

	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(codec)
	defer s.codecs.Remove(codec)

	c := initClient(codec, s.idgen, &s.services)
	<-codec.closed()
	c.Close()
}
