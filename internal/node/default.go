package node

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/adamnite/go-adamnite/internal/bargossip"
	"github.com/adamnite/go-adamnite/internal/config"
)

var DefaultConfig = Config{
	DataDir:        DefaultDataDir(),
	NodeType:       bargossip.NODE_TYPE_FULLNODE,
	P2PPort:        40908,
	BootstrapNodes: config.DefaultBootstrapNodes,
}

var DefaultBootNodeConfig = Config{
	DataDir:  DefaultDataDir(),
	NodeType: bargossip.NODE_TYPE_BOOTNODE,
	P2PPort:  40909,
}

func DefaultDataDir() string {
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
