package bargossip

import (
	"crypto/ecdsa"
	"io"
	"net"
	"sync"
	"time"
)

type peerTransportImpl struct {
	conn             net.Conn
	remotePeerPubKey *ecdsa.PublicKey
	rwmu             sync.RWMutex
}

func NewPeerTransport(conn net.Conn, remotePeerPubKey *ecdsa.PublicKey) peerTransport {
	return &peerTransportImpl{conn: conn, remotePeerPubKey: remotePeerPubKey}
}

// ********************************************************************************************
// ************************** Adamnite Transport Interface Functions **************************
// ********************************************************************************************

func (t *peerTransportImpl) doHandshake(prvKey *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	t.conn.SetDeadline(time.Now().Add(handshakeTimeout))

	var err error
	if t.remotePeerPubKey != nil {
		err = startHandshake(t.conn, prvKey, t.remotePeerPubKey)
	} else {
		err = receiveHandshake(t.conn, prvKey)
	}

	if err != nil {
		return nil, err
	}
}

func (t *peerTransportImpl) doExchangeProtocol() {

}

func (t *peerTransportImpl) close(err error) {
	t.rwmu.Lock()
	defer t.rwmu.Unlock()

	if t.conn != nil {

	}
	t.conn.Close()
}

func (t *peerTransportImpl) ReadMsg() (Msg, error) {

}

func (t *peerTransportImpl) WriteMsg(Msg) error {

}

// ********************************************************************************************
// ************************** Adamnite Transport internal Functions ***************************
// ********************************************************************************************

// startHandshake performs the handshake. This must be called on dial side.
func startHandshake(conn io.ReadWriter, prvKey *ecdsa.PrivateKey, remotePeerPubKey *ecdsa.PublicKey) error {

}

// receiveHandshake performs the handshake. This must be called on listening side
func receiveHandshake(conn io.ReadWriter, prvKey *ecdsa.PrivateKey) error {

}
