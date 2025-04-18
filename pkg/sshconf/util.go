package sshconf

import (
	"fmt"
	"os"
	"path/filepath"
)

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("etc", "ssh", "config"), nil
	}
	// home config
	path := filepath.Join(home, ".ssh", "config")
	if fileExists(path) {
		return path, nil
	}
	// server config
	path = filepath.Join("/", "etc", "ssh", "ssh_config")
	if fileExists(path) {
		return path, nil
	}
	return "", fmt.Errorf("unable to parse config %v: are you sure ssh is installed?", path)
}
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		// any other error we still return false
		return false
	}
	// file exists
	return true
}
