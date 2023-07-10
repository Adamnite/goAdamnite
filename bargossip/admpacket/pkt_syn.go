package admpacket

// SYN contains the handshake information
type SYN struct {
	Nonce Nonce
}

func (p *SYN) Name() string {
	return "ADM-SYN"
}

func (p *SYN) MessageType() byte {
	return SYNMsg
}

func (p *SYN) RequestID() []byte {
	return nil
}

func (p *SYN) SetRequestID(id []byte) {

}
