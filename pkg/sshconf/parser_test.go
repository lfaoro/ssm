// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package sshconf_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

func TestParse(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := sshconf.New()
		err := cfg.ParsePath("../../data/config_example")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Hosts) != 11 {
			t.Fatalf("expected 11 hosts, got %d", len(cfg.Hosts))
		}

		tests := []struct {
			name     string
			hostName string
			wantUser string
			wantHost string
			wantPort string
		}{
			{
				name:     "prod-api",
				hostName: "prod-api",
				wantUser: "deploy",
				wantHost: "api.example.com",
				wantPort: "22",
			},
			{
				name:     "terminalcoffee",
				hostName: "terminalcoffee",
				wantUser: "adam",
				wantHost: "terminal.shop",
				wantPort: "22",
			},
			{
				name:     "segfault.net",
				hostName: "segfault.net",
				wantUser: "root",
				wantHost: "segfault.net",
				wantPort: "22",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				host := cfg.GetHost(tt.hostName)
				if host.Name != tt.hostName {
					t.Errorf("host name = %q, want %q", host.Name, tt.hostName)
				}

				user, _ := host.Options.Get("user")
				if user != tt.wantUser {
					t.Errorf("user = %q, want %q", user, tt.wantUser)
				}

				hostname, _ := host.Options.Get("hostname")
				if hostname != tt.wantHost {
					t.Errorf("hostname = %q, want %q", hostname, tt.wantHost)
				}

				port, _ := host.Options.Get("port")
				if port != tt.wantPort {
					t.Errorf("port = %q, want %q", port, tt.wantPort)
				}
			})
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		cfg := sshconf.New()
		err := cfg.ParsePath("./nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestParseTags(t *testing.T) {
	t.Run("tags parsed", func(t *testing.T) {
		cfg := sshconf.New()
		err := cfg.ParsePath("../../data/config_example")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tests := []struct {
			name     string
			hostName string
			wantTags string
		}{
			{
				name:     "prod-api tags",
				hostName: "prod-api",
				wantTags: "production,api",
			},
			{
				name:     "terminalcoffee tags",
				hostName: "terminalcoffee",
				wantTags: "shops",
			},
			{
				name:     "dev-box tags",
				hostName: "dev-box",
				wantTags: "dev",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				host := cfg.GetHost(tt.hostName)
				tags, ok := host.Options.Get("#tag:")
				if !ok {
					t.Fatal("tag key not found")
				}
				if tags != tt.wantTags {
					t.Errorf("tags = %q, want %q", tags, tt.wantTags)
				}
			})
		}
	})
}

func TestTagOrder(t *testing.T) {
	t.Run("tag order mode", func(t *testing.T) {
		cfg := sshconf.New()
		cfg.SetOrder(sshconf.TagOrder)
		err := cfg.ParsePath("../../data/config_example")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Hosts) != 11 {
			t.Fatalf("expected 11 hosts, got %d", len(cfg.Hosts))
		}

		for i, h := range cfg.Hosts {
			_, hasTag := h.Options.Get("#tag:")
			if i < 8 && !hasTag {
				t.Errorf("expected host %q to have tags (tagged hosts first)", h.Name)
			}
			if i >= 8 && hasTag {
				t.Errorf("expected host %q to be untagged (untagged hosts last)", h.Name)
			}
		}
	})
}

func TestGetParamFor(t *testing.T) {
	t.Run("known host", func(t *testing.T) {
		cfg := sshconf.New()
		err := cfg.ParsePath("../../data/config_example")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		host := cfg.GetHost("staging-web")
		port := cfg.GetParamFor(host, "port")
		if port != "2222" {
			t.Errorf("port = %q, want %q", port, "2222")
		}
	})

	t.Run("unknown key", func(t *testing.T) {
		cfg := sshconf.New()
		err := cfg.ParsePath("../../data/config_example")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		host := cfg.GetHost("staging-web")
		val := cfg.GetParamFor(host, "nonexistent")
		if val != "" {
			t.Errorf("expected empty string for unknown key, got %q", val)
		}
	})

	t.Run("unknown host", func(t *testing.T) {
		cfg := sshconf.New()
		err := cfg.ParsePath("../../data/config_example")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		host := sshconf.Host{Name: "nobody"}
		val := cfg.GetParamFor(host, "port")
		if val != "" {
			t.Errorf("expected empty string for unknown host, got %q", val)
		}
	})
}

func TestGetHost(t *testing.T) {
	cfg := sshconf.New()
	err := cfg.ParsePath("../../data/config_example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("returns matching host", func(t *testing.T) {
		host := cfg.GetHost("prod-api")
		if host.Name != "prod-api" {
			t.Errorf("host.Name = %q, want %q", host.Name, "prod-api")
		}
	})

	t.Run("returns empty host for unknown name", func(t *testing.T) {
		host := cfg.GetHost("nonexistent")
		if host.Name != "" {
			t.Errorf("expected empty host, got %q", host.Name)
		}
	})
}

func TestGetHosts(t *testing.T) {
	cfg := sshconf.New()
	err := cfg.ParsePath("../../data/config_example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hosts := cfg.GetHosts()
	if len(hosts) != 11 {
		t.Fatalf("expected 11 hosts, got %d", len(hosts))
	}

	// verify it's a copy, not the same slice
	hosts[0].Name = "modified"
	original := cfg.GetHost("prod-api")
	if original.Name == "modified" {
		t.Error("GetHosts should return a copy, mutations should not affect original")
	}
}

func TestGetPath(t *testing.T) {
	cfg := sshconf.New()
	err := cfg.ParsePath("../../data/config_example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := cfg.GetPath()
	if path == "" {
		t.Error("expected non-empty path after ParsePath")
	}
}

func TestIncludeDepthLimit(t *testing.T) {
	t.Run("exceeds max depth", func(t *testing.T) {
		dir := t.TempDir()
		numFiles := 12 // maxIncludeDepth is 10

		files := make([]string, numFiles)
		for i := range numFiles {
			files[i] = filepath.Join(dir, "depth_"+string(rune('0'+i)))
		}

		for i := numFiles - 1; i >= 0; i-- {
			var content string
			if i < numFiles-1 {
				content = "Include " + files[i+1] + "\n"
			}
			content += "Host test" + string(rune('0'+i)) + "\n  HostName localhost\n"
			if err := os.WriteFile(files[i], []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
		}

		cfg := sshconf.New()
		err := cfg.ParsePath(files[0])
		if err == nil {
			t.Fatal("expected error for exceeded include depth")
		}
	})

	t.Run("within limit", func(t *testing.T) {
		dir := t.TempDir()
		numFiles := 3

		files := make([]string, numFiles)
		for i := range numFiles {
			files[i] = filepath.Join(dir, "depth_"+string(rune('0'+i)))
		}

		for i := numFiles - 1; i >= 0; i-- {
			var content string
			if i < numFiles-1 {
				content = "Include " + files[i+1] + "\n"
			}
			content += "Host test" + string(rune('0'+i)) + "\n  HostName localhost\n"
			if err := os.WriteFile(files[i], []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
		}

		cfg := sshconf.New()
		err := cfg.ParsePath(files[0])
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.Hosts) < numFiles {
			t.Errorf("expected at least %d hosts, got %d", numFiles, len(cfg.Hosts))
		}
	})
}

func TestIncludeCycleDetection(t *testing.T) {
	dir := t.TempDir()

	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")

	mustWrite(t, a, "Include "+b+"\nHost a\n  HostName a.local\n")
	mustWrite(t, b, "Include "+a+"\nHost b\n  HostName b.local\n")

	cfg := sshconf.New()
	err := cfg.ParsePath(a)
	if err == nil {
		t.Fatal("expected error for cyclic include")
	}
}

func TestIncludeGlob(t *testing.T) {
	dir := t.TempDir()

	mustWrite(t, filepath.Join(dir, "inc_a"), "Host a\n  HostName a.local\n")
	mustWrite(t, filepath.Join(dir, "inc_b"), "Host b\n  HostName b.local\n")

	main := filepath.Join(dir, "main")
	mustWrite(t, main, "Include "+filepath.Join(dir, "inc_*")+"\nHost main\n  HostName main.local\n")

	cfg := sshconf.New()
	err := cfg.ParsePath(main)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(cfg.Hosts))
	}
}

func TestIncludeRelativePath(t *testing.T) {
	dir := t.TempDir()

	mustWrite(t, filepath.Join(dir, "included"), "Host sub\n  HostName sub.local\n")
	mustWrite(t, filepath.Join(dir, "main"), "Include included\nHost main\n  HostName main.local\n")

	cfg := sshconf.New()
	err := cfg.ParsePath(filepath.Join(dir, "main"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(cfg.Hosts))
	}
}

func TestIncludePathTraversal(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0700); err != nil {
		t.Fatal(err)
	}

	mustWrite(t, filepath.Join(sub, "safe"), "Host safe\n  HostName safe.local\n")
	main := filepath.Join(dir, "main")
	mustWrite(t, main, "Include sub/../sub/safe\nHost main\n  HostName main.local\n")

	cfg := sshconf.New()
	err := cfg.ParsePath(main)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(cfg.Hosts))
	}
}

func TestParseDefaultPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	sshDir := filepath.Join(dir, ".ssh")
	if err := os.Mkdir(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(sshDir, "config"), "Host home\n  HostName home.local\n")

	cfg := sshconf.New()
	err := cfg.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(cfg.Hosts))
	}
	host := cfg.GetHost("home")
	if host.Name != "home" {
		t.Errorf("host name = %q, want %q", host.Name, "home")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
