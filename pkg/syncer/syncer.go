// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

// Package syncer orchestrates cloud provider server discovery and writes
// SSH config entries to a dedicated include file.
package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lfaoro/ssm/pkg/providers"
)

const (
	managedFileName = "50-ssm-cloud"
	includeLine     = "Include ~/.ssh/config.d/*"
)

var allProviders = []providers.Provider{
	providers.Hetzner{},
	providers.AWS{},
	providers.GCP{},
	providers.Azure{},
}

type Syncer struct {
	outputDir string
	sshConfig string
}

func New() *Syncer {
	return &Syncer{
		outputDir: sshConfigDir(),
		sshConfig: sshConfigPath(),
	}
}

func (s *Syncer) Sync(ctx context.Context, user, keyPath string, providerNames []string) ([]providers.Server, error) {
	provs := s.filterProviders(providerNames)
	var all []providers.Server
	for _, p := range provs {
		servers, err := p.FetchServers(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name(), err)
		}
		all = append(all, servers...)
	}
	if len(all) == 0 {
		return nil, nil
	}

	content := generateSSHConfig(all, user, keyPath)
	if err := writeManagedFile(s.outputDir, content); err != nil {
		return nil, fmt.Errorf("writing managed config: %w", err)
	}
	if err := ensureInclude(s.sshConfig); err != nil {
		return nil, fmt.Errorf("ensuring include directive: %w", err)
	}
	return all, nil
}

func (s *Syncer) DryRun(ctx context.Context, user, keyPath string, providerNames []string) (string, error) {
	provs := s.filterProviders(providerNames)
	var all []providers.Server
	for _, p := range provs {
		servers, err := p.FetchServers(ctx)
		if err != nil {
			return "", fmt.Errorf("%s: %w", p.Name(), err)
		}
		all = append(all, servers...)
		sort.Slice(all, func(i, j int) bool {
			if all[i].Provider != all[j].Provider {
				return all[i].Provider < all[j].Provider
			}
			return all[i].Name < all[j].Name
		})
	}
	return generateSSHConfig(all, user, keyPath), nil
}

func (s *Syncer) filterProviders(names []string) []providers.Provider {
	if len(names) == 0 {
		return allProviders
	}
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[strings.ToLower(n)] = true
	}
	var out []providers.Provider
	for _, p := range allProviders {
		if set[p.Name()] {
			out = append(out, p)
		}
	}
	return out
}

func generateSSHConfig(servers []providers.Server, user, keyPath string) string {
	sort.Slice(servers, func(i, j int) bool {
		if servers[i].Provider != servers[j].Provider {
			return servers[i].Provider < servers[j].Provider
		}
		return servers[i].Name < servers[j].Name
	})

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# SSM managed block - %s\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString("# Do not edit manually - changes will be overwritten by `ssm sync`\n")

	for _, s := range servers {
		hostname := sanitizeHostName(fmt.Sprintf("%s-%s", s.Region, s.Name))
		ip := s.PublicIP
		if ip == "" {
			ip = s.PrivateIP
		}
		if ip == "" {
			continue
		}

		b.WriteString(fmt.Sprintf("\nHost %s\n", hostname))
		b.WriteString(fmt.Sprintf("    HostName %s\n", ip))
		if user != "" {
			b.WriteString(fmt.Sprintf("    User %s\n", user))
		}
		if keyPath != "" {
			b.WriteString(fmt.Sprintf("    IdentityFile %s\n", keyPath))
		}
		b.WriteString(fmt.Sprintf("    #tag: %s\n", s.Provider))
	}
	b.WriteString("\n# End SSM managed block\n")
	return b.String()
}

func sanitizeHostName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if r == '_' || r == '.' {
			b.WriteRune('-')
		}
	}
	return b.String()
}

func writeManagedFile(dir, content string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, managedFileName)
	return os.WriteFile(path, []byte(content), 0o600)
}

func ensureInclude(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(configPath, []byte(includeLine+"\n"), 0o600)
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, includeLine) {
			return nil
		}
	}
	content := includeLine + "\n" + string(data)
	return os.WriteFile(configPath, []byte(content), 0o600)
}

func sshConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/", "etc", "ssh", "config.d")
	}
	return filepath.Join(home, ".ssh", "config.d")
}

func sshConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/", "etc", "ssh", "ssh_config")
	}
	return filepath.Join(home, ".ssh", "config")
}
