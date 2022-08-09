package rpc

import (
	"context"
)

type API struct {
	Namespace string
	Public    bool
	Version   string
	Service   interface{}
}

type jsonWriter interface {
	writeJSON(context.Context, interface{}) error
	closed() <-chan interface{}
	remoteAddr() string
}

type ServerCodec interface {
	readBatch() (msgs []*jsonrpcMessage, isBatch bool, err error)
	close()
	jsonWriter
}
