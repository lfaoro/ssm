// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

package sshconf_test

import (
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
