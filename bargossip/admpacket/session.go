package admpacket

type session struct {
	readkey      []byte
	writekey     []byte
	nonceCounter uint32
}

func (s *session) copy() *session {
	return &session{s.writekey, s.readkey, s.nonceCounter}
}
