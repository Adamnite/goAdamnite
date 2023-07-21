package admpacket

import (
	crand "crypto/rand"
	"encoding/binary"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/utils/mclock"
	"github.com/hashicorp/golang-lru/simplelru"
)

type sessionID struct {
	id     admnode.NodeID
	ipAddr string
}

type SessionCache struct {
	sessions   *simplelru.LRU
	clock      mclock.Clock
	handshakes map[sessionID]*AskHandshake
}

func NewSessionCache(maxItems int, clock mclock.Clock) *SessionCache {
	cache, err := simplelru.NewLRU(maxItems, nil)
	if err != nil {
		panic("cannot create session cache")
	}

	return &SessionCache{
		sessions:   cache,
		clock:      clock,
		handshakes: make(map[sessionID]*AskHandshake),
	}
}

func (sc *SessionCache) nextNonce(s *session) (Nonce, error) {
	s.nonceCounter++
	return generateNonce(s.nonceCounter)
}

// session returns the current session for the given node, if any.
func (sc *SessionCache) session(id admnode.NodeID, ipAddr string) *session {
	item, ok := sc.sessions.Get(sessionID{id, ipAddr})
	if !ok {
		return nil
	}
	return item.(*session)
}

// readKey returns the current read key for the given node.
func (sc *SessionCache) readKey(id admnode.NodeID, ipAddr string) []byte {
	if s := sc.session(id, ipAddr); s != nil {
		return s.readkey
	}
	return nil
}

// storeNewSession stores new encryption keys in the cache.
func (sc *SessionCache) storeNewSession(id admnode.NodeID, ipAddr string, s *session) {
	sc.sessions.Add(sessionID{id, ipAddr}, s)
}

func (sc *SessionCache) storeSentHandshake(id admnode.NodeID, ipAddr string, handshake *AskHandshake) {
	handshake.sent = sc.clock.Now()
	sc.handshakes[sessionID{id, ipAddr}] = handshake
}

func (sc *SessionCache) deleteHandshake(id admnode.NodeID, ipAddr string) {
	delete(sc.handshakes, sessionID{id, ipAddr})
}

func (sc *SessionCache) getHandshake(id admnode.NodeID, ipAddr string) *AskHandshake {
	return sc.handshakes[sessionID{id, ipAddr}]
}

func (sc *SessionCache) cleanHandshake() {
	deadline := sc.clock.Now().Add(-handshakeTimeout)
	for key, handshake := range sc.handshakes {
		if handshake.sent < deadline {
			delete(sc.handshakes, key)
		}
	}
}

func generateNonce(counter uint32) (n Nonce, err error) {
	binary.BigEndian.PutUint32(n[:4], counter)
	_, err = crand.Read(n[4:])
	return n, err
}

func generateMaskingIV(buf []byte) error {
	_, err := crand.Read(buf)
	return err
}
