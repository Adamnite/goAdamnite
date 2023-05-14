package node

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adamnite/go-adamnite/bargossip"
	"github.com/adamnite/go-adamnite/bargossip/nat"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/log15"
)

const (
	datadirPrivateKey      = "nodekey"
	datadirNodeDatabase    = "nodes"
	datadirDefaultKeystore = "keys"
)

type Config struct {
	Name        string `toml:"-"`
	Version     string `toml:"-"`
	DataDir     string
	KeyStoreDir string `toml:",omitempty"`
	IPCPath     string
	Logger      log15.Logger `toml:",omitempty"`
	P2P         bargossip.Config
}

var DefaultConfig = Config{
	Name:    "gnite",
	IPCPath: "gnite.ipc",
	DataDir: DefaultDataDir(),
	P2P: bargossip.Config{
		MaxPendingConnections:  50,
		MaxInboundConnections:  50,
		MaxOutboundConnections: 50,
		ListenAddr:             ":30900",
		NAT:                    nat.Any(),
	},
}

var DefaultDemoConfig = Config{
	Name:    "gnite-demo",
	IPCPath: "gnite-demo.ipc",
	DataDir: "",
	P2P: bargossip.Config{
		MaxPendingConnections:  50,
		MaxInboundConnections:  50,
		MaxOutboundConnections: 50,
		ListenAddr:             ":30900",
		NAT:                    nat.Any(),
	},
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Adamnite")
		case "windows":
			// We used to put everything in %HOME%\AppData\Roaming, but this caused
			// problems with non-typical setups. If this fallback location exists and
			// is non-empty, use it, otherwise DTRT and check %LOCALAPPDATA%.
			fallback := filepath.Join(home, "AppData", "Roaming", "Adamnite")
			appdata := windowsAppData()
			if appdata == "" || isNonEmptyDir(fallback) {
				return fallback
			}
			return filepath.Join(appdata, "Adamnite")
		default:
			return filepath.Join(home, ".adamnite")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func windowsAppData() string {
	v := os.Getenv("LOCALAPPDATA")
	if v == "" {
		// Windows XP and below don't have LocalAppData. Crash here because
		// we don't support Windows XP and undefining the variable will cause
		// other issues.
		panic("environment variable LocalAppData is undefined")
	}
	return v
}

func isNonEmptyDir(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	names, _ := f.Readdir(1)
	f.Close()
	return len(names) > 0
}

func (c *Config) IPCEndpoint() string {
	// Short circuit if IPC has not been enabled
	if c.IPCPath == "" {
		return ""
	}
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(c.IPCPath, `\\.\pipe\`) {
			return c.IPCPath
		}
		return `\\.\pipe\` + c.IPCPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(c.IPCPath) == c.IPCPath {
		if c.DataDir == "" {
			return filepath.Join(os.TempDir(), c.IPCPath)
		}
		return filepath.Join(c.DataDir, c.IPCPath)
	}
	return c.IPCPath
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
			log15.Crit(fmt.Sprintf("Failed to generate node key: %v", err))
		}
		return key
	}

	keyfile := c.ResolvePath(datadirPrivateKey)
	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		log15.Crit(fmt.Sprintf("Failed to generate node key: %v", err))
	}

	instanceDir := filepath.Join(c.DataDir, c.name())
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log15.Error(fmt.Sprintf("Failed to persist node key: %v", err))
		return key
	}

	keyfile = filepath.Join(instanceDir, datadirPrivateKey)
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		log15.Error(fmt.Sprintf("Failed to persist node key: %v", err))
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

func (c *Config) AccountConfig() (string, error) {
	var (
		keydir string
		err    error
	)

	switch {
	case filepath.IsAbs(c.KeyStoreDir):
		keydir = c.KeyStoreDir
	case c.DataDir != "":
		if c.KeyStoreDir == "" {
			keydir = filepath.Join(c.DataDir, datadirDefaultKeystore)
		} else {
			keydir, err = filepath.Abs(c.KeyStoreDir)
		}
	case c.KeyStoreDir != "":
		keydir, err = filepath.Abs(c.KeyStoreDir)
	}

	return keydir, err
}
