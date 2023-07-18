package bargossip

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/adamnite/go-adamnite/bargossip/nat"
	"github.com/adamnite/go-adamnite/crypto"
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
		NAT:                    natImpl,
	}

	server := &Server{
		Config: config,
	}
	if err := server.Start(); err != nil {
		t.Fatalf("could not start server: %v", err)
	}
	fmt.Println("server started")
	return server
}

func TestServer(t *testing.T) {
	startTestServer(t)

}
