// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package sshconf

import (
	"fmt"
	"os"
	"path/filepath"
)

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		//nolint:nilerr // fallback to system-wide config when $HOME is unavailable
		return filepath.Join("/", "etc", "ssh", "ssh_config"), nil
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

// RemoveComments strips inline comments from SSH config directives.
func RemoveComments(input string) string {
	return removeComments(input)
}

// IsSensitiveKey reports whether k is a sensitive SSH config key
// that should be hidden from the config viewport.
func IsSensitiveKey(k string) bool {
	return isSensitiveKey(k)
}
