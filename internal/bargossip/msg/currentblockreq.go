package msg

func NewCreateCurrentBlockRequest(requestID uint64) *BargossipMsg {
	message := &BargossipMsg{
		MsgType:   CurrentBlockRequest,
		RequestID: requestID,
		Length:    0,
		Data:      make([]byte, 0),
	}

	return message
}
