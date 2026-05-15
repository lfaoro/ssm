// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

// Package sshconf loads, parses SSH config files,
// tries to be thread-safe.
// ref: https://man.openbsd.org/ssh_config.5
package sshconf

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	som "github.com/thalesfsp/go-common-types/safeorderedmap"
)

// Config holds parsed SSH config data and provides thread-safe access.
type Config struct {
	mu             sync.Mutex
	Hosts          []Host
	secondaryHosts []Host

	order Order
	path  string
}

// Host represents a single SSH host entry with its options.
type Host struct {
	Name    string
	Options *som.SafeOrderedMap[string]
}

// Order defines host ordering strategies.
type Order int

const (
	// TagOrder prioritises hosts with #tag: comments.
	TagOrder Order = iota + 1
)

// New returns a new Config ready for parsing.
func New() *Config {
	return &Config{}
}

// SetOrder configures the host ordering strategy.
func (c *Config) SetOrder(o Order) {
	c.order = o
}

// Parse loads and parses the default SSH config file.
func (c *Config) Parse() error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	return c.parse(path, 0, nil)
}

// ParsePath loads and parses the SSH config file at the given path.
func (c *Config) ParsePath(s string) error {
	if !strings.HasPrefix(s, "/") {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		s = filepath.Join(wd, s)
	}
	resolved, err := filepath.EvalSymlinks(s)
	if err != nil {
		resolved = s
	}
	s = resolved
	return c.parse(s, 0, nil)
}

// GetHost returns the host entry matching the given name.
func (c *Config) GetHost(name string) Host {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, h := range c.Hosts {
		if h.Name == name {
			return h
		}
	}
	return Host{}
}

// GetParamFor returns the value of a config key for the given host.
func (c *Config) GetParamFor(host Host, key string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, h := range c.Hosts {
		if h.Name == host.Name {
			val, ok := h.Options.Get(key)
			if !ok {
				return ""
			}
			return val
		}
	}
	return ""
}

// GetHosts returns a copy of all parsed hosts.
func (c *Config) GetHosts() []Host {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Host, len(c.Hosts))
	copy(out, c.Hosts)
	return out
}

// GetPath returns the path of the parsed SSH config file.
func (c *Config) GetPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.path
}

const (
	commentPrefix  = "#"
	tagPrefix      = "#tag:"
	tagOrderPrefix = "#tagorder"

	maxIncludeDepth = 10
)

func (c *Config) parse(path string, depth int, visited map[string]bool) error {
	if depth > maxIncludeDepth {
		return fmt.Errorf("sshconf: max Include depth (%d) exceeded at %s", maxIncludeDepth, path)
	}
	if visited == nil {
		visited = make(map[string]bool)
	}
	absPath, err := filepath.Abs(path)
	if err == nil {
		if visited[absPath] {
			return fmt.Errorf("sshconf: cyclic Include detected: %s", path)
		}
		visited[absPath] = true
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.Hosts = []Host{}
	c.secondaryHosts = []Host{}

	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	info, err := f.Stat()
	if err == nil {
		if perm := info.Mode().Perm(); perm&0077 != 0 {
			fmt.Fprintf(os.Stderr, "ssm: warning: %s has insecure permissions %04o (should be 0600)\n", path, perm)
		}
	}
	c.path = path
	scanner := bufio.NewScanner(f)
	var tagOrder bool
	var currentHost *Host
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == tagOrderPrefix {
			tagOrder = true
		}
		if c.order == TagOrder {
			tagOrder = true
		}

		if line == "" ||
			strings.HasPrefix(line, commentPrefix) &&
				!strings.HasPrefix(line, tagPrefix) {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		k, v := strings.ToLower(parts[0]), strings.Join(parts[1:], " ")
		if !strings.HasPrefix(line, tagPrefix) {
			k = removeComments(k)
			v = removeComments(v)
		}
		if k == "include" {
			if !strings.HasPrefix(v, "/") {
				dir := filepath.Dir(path)
				v = filepath.Join(dir, v)
			}
			paths, err := filepath.Glob(v)
			if err != nil {
				return err
			}

			for _, incPath := range paths {
				cfg := New()
				err := cfg.parse(incPath, depth+1, visited)
				if err != nil {
					return err
				}
				c.Hosts = append(c.Hosts, cfg.Hosts...)
			}
		}
		if k == "host" {
			if strings.Contains(v, "*") {
				continue
			}
			if currentHost != nil {
				newHost(tagOrder, currentHost, c)
			}
			currentHost = &Host{
				Name:    v,
				Options: som.New[string](),
			}
			continue
		}
		if currentHost != nil {
			currentHost.Options.Add(k, v)
		}
	}
	if currentHost != nil {
		newHost(tagOrder, currentHost, c)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	c.Hosts = append(c.Hosts, c.secondaryHosts...)
	return nil
}

func newHost(tagOrder bool, currentHost *Host, config *Config) {
	if tagOrder {
		if currentHost.Options.Contains("#tag:") {
			config.Hosts = append(config.Hosts, *currentHost)
		} else {
			config.secondaryHosts = append(config.secondaryHosts, *currentHost)
		}
		return
	}
	config.Hosts = append(config.Hosts, *currentHost)
}

func removeComments(input string) string {
	if index := strings.Index(input, "#"); index != -1 {
		return strings.TrimSpace(input[:index])
	}
	return strings.TrimSpace(input)
}
