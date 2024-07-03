package node

type Config struct {
	Name     string `toml:"-"`
	Version  string `toml:"-"`
	P2PPort  uint32 `toml:"-"`
	DataDir  string
	NodeType uint8
}
