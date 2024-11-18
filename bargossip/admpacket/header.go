package admpacket

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
)

func (h *Header) mask(id admnode.NodeID) cipher.Stream {
	block, err := aes.NewCipher(id[:16])
	if err != nil {
		panic("cannot create cipher")
	}
	return cipher.NewCTR(block, h.IV[:])
}

func (h *Header) checkValid(packetLen int) error {
	if h.ProtocolID != admPacketProtocolID {
		return errInvalidHeader
	}

	if h.AuthSize > uint16(packetLen) {
		return errAuthSize
	}
	return nil
}
