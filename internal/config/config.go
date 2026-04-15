package config

import (
	"os"
	"path/filepath"
)

const EnvStoreDir = "WACLI_STORE_DIR"

func DefaultStoreDir() string {
	if dir := os.Getenv(EnvStoreDir); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".wacli"
	}
	return filepath.Join(home, ".wacli")
}
