// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
)

func resolvePingTarget(host sshconf.Host) (string, string) {
	hostname, _ := host.Options.Get("hostname")
	if hostname == "" {
		hostname = host.Name
	}
	port, _ := host.Options.Get("port")
	if port == "" {
		port = "22"
	}
	return hostname, port
}

func pingHost(hostname, port string) (time.Duration, error) {
	addr := net.JoinHostPort(hostname, port)
	start := time.Now()
	dialer := net.Dialer{Timeout: 20 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(start), nil
}

func isTimeoutErr(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func pingErrorLabel(err error) string {
	if err == nil {
		return ""
	}
	if isTimeoutErr(err) {
		return "timeout"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "unreachable"
	}
	return "err"
}

func pingLatency(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1 {
		return "<1ms"
	}
	return fmt.Sprintf("%dms", ms)
}

func pingSelectedCmd(m *Model) tea.Cmd {
	host, ok := m.li.SelectedItem().(item)
	if !ok {
		return AddLog("ping: no selected item")
	}
	h := m.config.GetHost(host.title)
	hostname, port := resolvePingTarget(h)
	return func() tea.Msg {
		latency, err := pingHost(hostname, port)
		if err != nil {
			return PingResultMsg{Host: host.title, Latency: pingErrorLabel(err)}
		}
		return PingResultMsg{Host: host.title, Latency: pingLatency(latency)}
	}
}

func pingAllCmd(m *Model) tea.Cmd {
	items := m.li.VisibleItems()
	if len(items) == 0 {
		return nil
	}

	workers := pingWorkerCount()
	sem := make(chan struct{}, workers)

	var cmds []tea.Cmd
	for _, it := range items {
		hostItem, ok := it.(item)
		if !ok {
			continue
		}
		hostName := hostItem.title
		h := m.config.GetHost(hostName)
		hostname, port := resolvePingTarget(h)

		cmds = append(cmds, func() tea.Msg {
			sem <- struct{}{}
			defer func() { <-sem }()

			latency, err := pingHost(hostname, port)
			if err != nil {
				return PingResultMsg{Host: hostName, Latency: pingErrorLabel(err)}
			}
			return PingResultMsg{Host: hostName, Latency: pingLatency(latency)}
		})
	}
	return tea.Batch(cmds...)
}

func refreshList(m *Model) {
	hosts := m.config.GetHosts()
	items := make([]list.Item, 0, len(hosts))
	for _, host := range hosts {
		items = append(items, formatHost(host, m.pingResults[host.Name]))
	}
	m.li.SetItems(items)

	if m.li.IsFiltered() {
		m.li.SetFilterText(m.li.FilterValue())
	}
}

// ConcurrencyLimit returns a safe, bounded number of concurrent workers
// suitable for batch operations (ping, remote command execution, etc.).
// It is intentionally conservative.
//
// We use CPU count as the main signal (cheap and portable) and apply hard
// bounds so we never launch a ridiculous number of goroutines even on
// high-core machines or containers that report high CPU counts.
func ConcurrencyLimit() int {
	return min(max(runtime.NumCPU()*4, 8), 64)
}

// pingWorkerCount is the old unexported name; kept for internal ping use
// during the transition. New code should use ConcurrencyLimit.
func pingWorkerCount() int { return ConcurrencyLimit() }
