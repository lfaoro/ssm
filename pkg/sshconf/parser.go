// sshconf loads, parses and watches SSH config files,
// tries to be thread-safe.
// ref: https://man.openbsd.org/ssh_config.5
package sshconf

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	// protects Hosts
	mu    sync.Mutex
	Hosts []Host
	path  string
	watch []string
}

type Host struct {
	Name    string
	Options map[string]string
}

// Parse parses SSH config files from default known locations.
// User: ~/.ssh/config
// System: /etc/ssh/ssh_config
// Parse also follows `Include` statements.
func Parse() (*Config, error) {
	path, err := defaultConfigPath()
	if err != nil {
		return &Config{}, err
	}
	return parse(path)
}

// Parse parses SSH config file from custom location.
func ParsePath(s string) (*Config, error) {
	return parse(s)
}

func (c *Config) GetHost(name string) Host {
	for _, h := range c.Hosts {
		if h.Name == name {
			return h
		}
	}
	return Host{}
}

func (c *Config) GetParamFor(host Host, key string) string {
	for _, h := range c.Hosts {
		if h.Name == host.Name {
			return h.Options[key]
		}
	}
	return ""
}

func (c *Config) GetPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.path
}

func (c *Config) Watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	for _, path := range c.watch {
		err = watcher.Add(path)
		if err != nil {
			return err
		}
	}
	for event := range watcher.Events {
		// fmt.Printf("detected event %v %v\n", event.Name, event.Op)
		if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) {
			// add path again to watcher in case editors rename the file
			// changing the inode
			err = watcher.Add(event.Name)
			if err != nil {
				return err
			}

			c.mu.Lock()
			// fmt.Println("reloading")
			conf, err := parse(c.path)
			if err != nil {
				c.mu.Unlock()
				return err
			}
			c.Hosts = conf.Hosts
			c.mu.Unlock()
		}
	}
	return nil
}

func parse(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	config := &Config{Hosts: []Host{}, path: path}
	scanner := bufio.NewScanner(f)
	var currentHost *Host
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// ignore empty or comment line
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		// malformed line, skip
		if len(parts) < 2 {
			continue
		}
		k, v := strings.ToLower(parts[0]), strings.Join(parts[1:], " ")
		// recurse include files
		if k == "include" {
			if !strings.HasPrefix(v, "/") {
				path = filepath.Dir(path)
				v = filepath.Join(path, v)
			}
			paths, err := filepath.Glob(v)
			if err != nil {
				return nil, err
			}

			for _, path := range paths {
				config.watch = append(config.watch, path)
				cfg, err := parse(path) // recursion
				if err != nil {
					return nil, err
				}
				config.Hosts = append(config.Hosts, cfg.Hosts...)
			}
		}
		// all blocks must start with Host key
		if k == "host" {
			if currentHost != nil {
				config.Hosts = append(config.Hosts, *currentHost)
			}
			currentHost = &Host{
				Name:    v,
				Options: map[string]string{},
			}
			continue
		}
		// if not a host key must be an option
		if currentHost != nil {
			currentHost.Options[k] = v
		}
	}
	if currentHost != nil {
		config.Hosts = append(config.Hosts, *currentHost)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}
