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
	mu             sync.RWMutex
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
	c.mu.Lock()
	c.order = o
	c.mu.Unlock()
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, h := range c.Hosts {
		if h.Name == name {
			return h
		}
	}
	return Host{}
}

// GetParamFor returns the value of a config key for the given host.
func (c *Config) GetParamFor(host Host, key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Host, len(c.Hosts))
	copy(out, c.Hosts)
	return out
}

// GetPath returns the path of the parsed SSH config file.
func (c *Config) GetPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
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

	// Capture order under a brief read lock so SetOrder cannot race with us.
	c.mu.RLock()
	order := c.order
	c.mu.RUnlock()

	// Perform all I/O and parsing without holding the lock.
	// Only take the lock at the very end to publish results atomically.

	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	info, statErr := f.Stat()
	if statErr == nil {
		if perm := info.Mode().Perm(); perm&0077 != 0 {
			fmt.Fprintf(os.Stderr, "ssm: warning: %s has insecure permissions %04o (should be 0600)\n", path, perm)
		}
	}

	scanner := bufio.NewScanner(f)
	var tagOrder bool
	if order == TagOrder {
		tagOrder = true
	}

	newHosts := []Host{}
	newSecondary := []Host{}

	var currentHost *Host
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == tagOrderPrefix {
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
				v = filepath.Clean(filepath.Join(dir, v))
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
				newHosts = append(newHosts, cfg.Hosts...)
			}
		}
		if k == "host" {
			if strings.Contains(v, "*") {
				continue
			}
			if currentHost != nil {
				appendHostToResults(&newHosts, &newSecondary, *currentHost, tagOrder)
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
		appendHostToResults(&newHosts, &newSecondary, *currentHost, tagOrder)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	newHosts = append(newHosts, newSecondary...)

	// Publish results under the lock (very short critical section)
	c.mu.Lock()
	c.Hosts = newHosts
	c.secondaryHosts = nil // not needed after append
	c.path = path
	c.mu.Unlock()

	return nil
}

func removeComments(input string) string {
	if before, _, ok := strings.Cut(input, "#"); ok {
		return strings.TrimSpace(before)
	}
	return strings.TrimSpace(input)
}

// appendHostToResults decides whether a host goes into the main list or the
// secondary (untagged) list, based on the current ordering mode.
// This is a small internal helper to avoid duplicating the decision logic.
func appendHostToResults(hosts, secondary *[]Host, h Host, tagOrder bool) {
	if tagOrder {
		if h.Options.Contains("#tag:") {
			*hosts = append(*hosts, h)
		} else {
			*secondary = append(*secondary, h)
		}
	} else {
		*hosts = append(*hosts, h)
	}
}

var sensitiveKeys = map[string]bool{
	"identityfile":         true,
	"certificatefile":      true,
	"proxycommand":         true,
	"pkcs11provider":       true,
	"controlpath":          true,
	"userknownhostsfile":   true,
	"revokedhostkeys":      true,
	"globalknownhostsfile": true,
}

func isSensitiveKey(k string) bool {
	return sensitiveKeys[k]
}
