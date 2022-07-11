package bargossip

import (
	"crypto/ecdsa"
	"testing"

	"github.com/adamnite/go-adamnite/bargossip/nat"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/internal/testlog"
	"github.com/adamnite/go-adamnite/log15"
)

func genKey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}

func startTestServer(t *testing.T) *Server {
	natImpl, _ := nat.Parse("extip:192.168.112.109")
	config := Config{
		ServerPrvKey:           genKey(),
		ListenAddr:             "127.0.0.1:8000",
		MaxInboundConnections:  10,
		MaxOutboundConnections: 10,
		Logger:                 testlog.Logger(t, log15.LvlTrace),
		NAT:                    natImpl,
	}

	server := &Server{
		Config: config,
	}
	if err := server.Start(); err != nil {
		t.Fatalf("could not start server: %v", err)
	}
	return server
}

func TestServer(t *testing.T) {
	startTestServer(t)
}
