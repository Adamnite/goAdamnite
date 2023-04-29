package rpc

import (
	"github.com/vmihailenco/msgpack/v5"
)

func Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func Decode(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}
