package xdg

import (
	"os"
	"path/filepath"
)

func OpenCodeData() (string, bool) {
	if v := os.Getenv("OPENCODE_DATA"); v != "" {
		return v, true
	}
	return "", false
}

func DataHome() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v
	}
	return filepath.Join(os.Getenv("HOME"), ".local", "share")
}

func ConfigHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	return filepath.Join(os.Getenv("HOME"), ".config")
}

func DefaultDB() string {
	return filepath.Join(DataHome(), "opencode", "opencode.db")
}
