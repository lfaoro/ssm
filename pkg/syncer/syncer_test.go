// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package syncer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lfaoro/ssm/pkg/providers"
)

func TestGenerateSSHConfig(t *testing.T) {
	tests := []struct {
		name    string
		servers []providers.Server
		user    string
		keyPath string
		checks  []string
		noMatch []string
	}{
		{
			name: "single server no user no key",
			servers: []providers.Server{
				{Name: "web-01", PublicIP: "1.2.3.4", Provider: "hetzner", Region: "fsn1"},
			},
			checks: []string{
				"Host fsn1-web-01",
				"HostName 1.2.3.4",
				"#tag: hetzner",
			},
			noMatch: []string{"User", "IdentityFile"},
		},
		{
			name: "with user and key",
			servers: []providers.Server{
				{Name: "db-01", PublicIP: "5.6.7.8", Provider: "aws", Region: "eu-west-1"},
			},
			user:    "deploy",
			keyPath: "~/.ssh/id_ed25519",
			checks: []string{
				"User deploy",
				"IdentityFile ~/.ssh/id_ed25519",
			},
		},
		{
			name: "no ip skipped",
			servers: []providers.Server{
				{Name: "hidden", Provider: "gcp", Region: "us-central1"},
			},
			checks:  nil,
			noMatch: []string{"Host hidden"},
		},
		{
			name: "private ip fallback",
			servers: []providers.Server{
				{Name: "internal", PrivateIP: "10.0.0.1", Provider: "azure", Region: "eastus"},
			},
			checks: []string{"HostName 10.0.0.1"},
		},
		{
			name: "multiple servers sorted by name",
			servers: []providers.Server{
				{Name: "zeta", PublicIP: "3.3.3.3", Provider: "hetzner", Region: "fsn1"},
				{Name: "alpha", PublicIP: "1.1.1.1", Provider: "hetzner", Region: "fsn1"},
			},
			checks: []string{"Host fsn1-alpha", "Host fsn1-zeta"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := generateSSHConfig(tt.servers, tt.user, tt.keyPath)
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected output to contain %q", c)
				}
			}
			for _, n := range tt.noMatch {
				if strings.Contains(out, n) {
					t.Errorf("expected output NOT to contain %q", n)
				}
			}
			if !strings.HasPrefix(out, "# SSM managed block") {
				t.Error("expected header comment")
			}
			if !strings.HasSuffix(strings.TrimSpace(out), "# End SSM managed block") {
				t.Error("expected footer comment")
			}
		})
	}
}

func TestSanitizeHostName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My.Server_01", "my-server-01"},
		{"foo@bar!", "foobar"},
		{"UPPERCASE", "uppercase"},
		{"---", "---"},
		{"abc123", "abc123"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeHostName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeHostName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureInclude(t *testing.T) {
	t.Run("creates new file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config")
		if err := ensureInclude(path); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), includeLine) {
			t.Errorf("expected include line %q in output", includeLine)
		}
	})

	t.Run("no duplicate when already present", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config")
		content := includeLine + "\nHost test\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := ensureInclude(path); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), includeLine) != 1 {
			t.Errorf("expected exactly one include line, got:\n%s", data)
		}
	})

	t.Run("prepends when missing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config")
		content := "Host test\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := ensureInclude(path); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), includeLine) {
			t.Errorf("expected include line at start, got:\n%s", data)
		}
	})
}

func TestWriteManagedFile(t *testing.T) {
	t.Run("creates dirs and writes file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sub", "50-ssm-test")
		content := "Host test\n"
		if err := writeManagedFile(path, content); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path) //nolint:gosec
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != content {
			t.Errorf("got %q, want %q", data, content)
		}
	})

	t.Run("writes with 0600 permissions", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "50-ssm-test")
		if err := writeManagedFile(path, "data"); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("expected 0600, got %04o", perm)
		}
	})
}

func TestPath(t *testing.T) {
	s := New()
	got := s.Path("hetzner")
	if !strings.HasSuffix(got, "/config.d/50-ssm-hetzner") {
		t.Errorf("unexpected path: %s", got)
	}
}
