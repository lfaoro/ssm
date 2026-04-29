// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

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

type Config struct {
	mu             sync.Mutex
	Hosts          []Host
	secondaryHosts []Host

	order Order
	path  string
}

type Host struct {
	Name    string
	Options *som.SafeOrderedMap[string]
}

type Order int

const (
	TagOrder Order = iota + 1
)

func New() *Config {
	return &Config{}
}

func (c *Config) SetOrder(o Order) {
	c.order = o
}

func (c *Config) Parse() error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	return c.parse(path, 0, nil)
}

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

func (c *Config) GetHosts() []Host {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Host, len(c.Hosts))
	copy(out, c.Hosts)
	return out
}

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

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

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
