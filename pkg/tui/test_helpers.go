// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

func testConfigPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join("testdata", "test_config")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}
	return abs
}

func newTestConfig(t *testing.T) *sshconf.Config {
	t.Helper()
	cfg := sshconf.New()
	err := cfg.ParsePath(testConfigPath(t))
	if err != nil {
		t.Fatalf("failed to parse test config: %v", err)
	}
	return cfg
}

func newTestModel(t *testing.T, debug bool) *Model {
	t.Helper()
	cfg := newTestConfig(t)
	m := NewModel(cfg, debug, SSHCmd)
	m.li.SetSize(80, 24)
	return m
}

func skipIfCmdNotFound(t *testing.T, cmd string) {
	t.Helper()
	if _, err := exec.LookPath(cmd); err != nil {
		t.Skipf("skipping: %q not found in PATH", cmd)
	}
}

func skipIfNoEditor(t *testing.T) {
	t.Helper()
	if os.Getenv("EDITOR") != "" {
		return
	}
	for _, editor := range []string{"vim", "vi", "nano", "ed"} {
		if _, err := exec.LookPath(editor); err == nil {
			return
		}
	}
	t.Skip("skipping: no editor found")
}
