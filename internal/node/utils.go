package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func checkNodeKey(config *Config) crypto.PrivKey {
	var nodeFileName string

	switch config.NodeType {
	case NODE_TYPE_BOOTNODE:
		nodeFileName = "bootnode"
	case NODE_TYPE_FULLNODE:
		nodeFileName = "gnite"
	}

	keyPath := filepath.Join(config.DataDir, nodeFileName, "node.key")

	if fileExists(keyPath) {
		// file read
		prvKey, err := readNodeKey(config)
		if err != nil {
			return nil
		}

		return prvKey
	} else {
		return nil
	}
}

func saveNodeKey(config *Config, prvKey crypto.PrivKey) error {
	var nodeFileName string

	switch config.NodeType {
	case NODE_TYPE_BOOTNODE:
		nodeFileName = "bootnode"
	case NODE_TYPE_FULLNODE:
		nodeFileName = "gnite"
	}

	nodeKeyPath := filepath.Join(config.DataDir, nodeFileName, "node.key")

	if fileExists(config.DataDir) {

	} else {
		err := os.MkdirAll(filepath.Join(config.DataDir, nodeFileName), 0755)
		if err != nil {
			return err
		}
		byPrvKey, err := crypto.MarshalPrivateKey(prvKey)
		if err != nil {
			return err
		}

		err = fileSave(nodeKeyPath, byPrvKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func readNodeKey(config *Config) (crypto.PrivKey, error) {
	var nodeFileName string

	switch config.NodeType {
	case NODE_TYPE_BOOTNODE:
		nodeFileName = "bootnode"
	case NODE_TYPE_FULLNODE:
		nodeFileName = "gnite"
	}

	nodeKeyPath := filepath.Join(config.DataDir, nodeFileName, "node.key")

	if fileExists(nodeKeyPath) {
		content, err := os.ReadFile(nodeKeyPath)
		if err != nil {
			return nil, err
		}

		prvKey, err := crypto.UnmarshalPrivateKey(content)
		if err != nil {
			return nil, err
		}
		return prvKey, nil
	} else {
		return nil, fmt.Errorf("node.key not exist")
	}
}

// fileExists checks if a file exists and is not a directory before we try using it
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func fileSave(filename string, contents []byte) error {
	// Create or truncate the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write some content to the file
	_, err = file.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

func fileSaveWithString(filename string, contents string) error {
	// Create or truncate the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write some content to the file
	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}

	return nil
}
