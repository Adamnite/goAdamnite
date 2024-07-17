package node

import (
	maddr "github.com/multiformats/go-multiaddr"
)

type Config struct {
	Name           string `toml:"-"`
	Version        string `toml:"-"`
	P2PPort        uint32 `toml:"-"`
	DataDir        string
	NodeType       uint8
	BootstrapNodes []maddr.Multiaddr
	ProtocolID     string
}
