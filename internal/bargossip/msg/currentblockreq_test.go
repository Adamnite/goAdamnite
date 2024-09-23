package msg

import "testing"

func TestSerializeAndDeserialize(t *testing.T) {
	currentBlockRequest := NewCreateCurrentBlockRequest(1)

	bytes, err := currentBlockRequest.ToSerialize()
	if err != nil {
		t.Error("Cannot serialize")
	}

	emptyMst := &BargossipMsg{}
	err = emptyMst.ToDeserialize(bytes)
	if err != nil {
		t.Error("Cannot deserialize")
	}

	if currentBlockRequest.RequestID == emptyMst.RequestID && currentBlockRequest.Length == emptyMst.Length && currentBlockRequest.MsgType == emptyMst.MsgType {

	} else {
		t.Error("Failed")
	}
}
