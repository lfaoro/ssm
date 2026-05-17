// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"errors"
	"fmt"
	"net"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
)

func resolvePingTarget(host sshconf.Host) (hostname, port string) {
	hostname, _ = host.Options.Get("hostname")
	if hostname == "" {
		hostname = host.Name
	}
	port, _ = host.Options.Get("port")
	if port == "" {
		port = "22"
	}
	return
}

func pingHost(hostname, port string) (time.Duration, error) {
	addr := net.JoinHostPort(hostname, port)
	start := time.Now()
	dialer := net.Dialer{Timeout: 3 * time.Second}
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
	hosts := m.config.GetHosts()
	cmds := make([]tea.Cmd, 0, len(hosts))
	for _, host := range hosts {
		hostname, port := resolvePingTarget(host)
		hostName := host.Name
		cmds = append(cmds, func() tea.Msg {
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
