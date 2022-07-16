package admpacket

import "errors"

var (
	errTooShort          = errors.New("packet too short")
	errInvalidHeader     = errors.New("invalid packet header")
	errAuthSize          = errors.New("declared auth size is beyond packet length")
	errInvalidPacketType = errors.New("invalid packet type in header")
	errMessageDecrypt    = errors.New("cannot decrypt message")
	errMessageTooShort   = errors.New("packet message too short")
	ErrInvalidReqID      = errors.New("request ID larger than 8 bytes")
	errNoRecord          = errors.New("expected nodeinfo in handshake but none sent")
	errInvalidSig        = errors.New("invalid signature")
)
