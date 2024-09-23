package msg

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

const (
	CurrentBlockRequest  = 1
	CurrentBlockResponse = 2
)

type BargossipMsg struct {
	MsgType   uint
	RequestID uint64
	Length    uint
	Data      []byte
}

func (msg *BargossipMsg) ToSerialize() ([]byte, error) {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(msg); err != nil {
		fmt.Println("Error serializing to binary:", err)
		return nil, err
	}

	binaryBytes := buffer.Bytes()
	return binaryBytes, nil
}

func (msg *BargossipMsg) ToDeserialize(binaryBytes []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(binaryBytes))
	if err := decoder.Decode(msg); err != nil {
		fmt.Println("Error decoding from binary:", err)
		return err
	}
	return nil
}
