package config

import (
	"log"

	"github.com/multiformats/go-multiaddr"
)

var (
	DefaultBootstrapNodes []multiaddr.Multiaddr

	DefualtBootstrapNodesFromString = []string{
		"/ip4/127.0.0.1/tcp/40909/p2p/QmVkCj3HhaVmao2NjDBJ9ZZcZkEZHLaa5D56gj2Kw4nKKu"}
)

func init() {
	for _, str := range DefualtBootstrapNodesFromString {
		ma, err := multiaddr.NewMultiaddr(str)
		if err != nil {
			log.Fatal("Default bootstrap node string is unknown format!!!")
		}

		DefaultBootstrapNodes = append(DefaultBootstrapNodes, ma)
	}
}
