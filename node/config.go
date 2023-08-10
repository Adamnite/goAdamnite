package node

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/crypto"

	log "github.com/sirupsen/logrus"
)

const (
	datadirPrivateKey      = "nodekey"
	datadirNodeDatabase    = "nodes"
	datadirDefaultKeystore = "keys"
)

type Config struct {
	Name        string
	Version     string
	DataDir     string
	KeyStoreDir string
	IPCPath     string
	P2P         bargossip.Config
}

func (c *Config) name() string {
	if c.Name == "" {
		progname := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if progname == "" {
			panic("empty executable name, set Config.Name")
		}
		return progname
	}
	return c.Name
}

func (c *Config) NodeName() string {
	name := c.name()
	if name == "gnite" {
		name = "GNite"
	}

	if c.Version != "" {
		name += "/v" + c.Version
	}

	name += "/" + runtime.GOOS + "-" + runtime.GOARCH
	name += "/" + runtime.Version()
	return name
}

// NodeKey retrieves the currently configured private key of the node
func (c *Config) NodeKey() *ecdsa.PrivateKey {
	if c.P2P.ServerPrvKey != nil {
		return c.P2P.ServerPrvKey
	}

	if c.DataDir == "" {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatal(fmt.Sprintf("Failed to generate node key: %v", err))
		}
		return key
	}

	keyfile := c.ResolvePath(datadirPrivateKey)
	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to generate node key: %v", err))
	}

	instanceDir := filepath.Join(c.DataDir, c.name())
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return key
	}

	keyfile = filepath.Join(instanceDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log.Error(fmt.Sprintf("Failed to persist node key: %v", err))
	}
	return key
}

func (c *Config) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if c.DataDir == "" {
		return ""
	}

	return filepath.Join(c.instanceDir(), path)
}

func (c *Config) instanceDir() string {
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, c.name())
}

func (c *Config) NodeDB() string {
	if c.DataDir == "" {
		return "" // ephemeral
	}
	return c.ResolvePath(datadirNodeDatabase)
}
