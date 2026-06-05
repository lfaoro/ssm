// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"strings"
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

func BenchmarkSetConfig(b *testing.B) {
	cfg := sshconf.New()
	if err := cfg.ParsePath("../../data/config_example"); err != nil {
		b.Fatalf("ParsePath: %v", err)
	}
	m := NewModel(cfg, false, SSHCmd)
	m.li.SetSize(80, 24)
	b.ReportAllocs()
	for b.Loop() {
		m.setConfig()
	}
}

func BenchmarkFormatHost(b *testing.B) {
	cfg := sshconf.New()
	if err := cfg.ParsePath("../../data/config_example"); err != nil {
		b.Fatalf("ParsePath: %v", err)
	}
	hosts := cfg.GetHosts()
	b.ReportAllocs()
	for b.Loop() {
		for _, h := range hosts {
			_ = formatHost(h)
		}
	}
}

func BenchmarkSanitizeOutput(b *testing.B) {
	inputs := []struct {
		name  string
		value string
	}{
		{"clean", "hello world"},
		{"ansi_colors", "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m"},
		{"cursor_moves", "\x1b[2J\x1b[H\x1b[?25l"},
		{"hyperlink", "\x1b]8;;https://example.com\x07link\x1b]8;;\x07"},
		{"mixed", "\x1b[1mbold\x1b[0m\n\x1b[33mwarn\x1b[0m\r\x1b[2K"},
	}
	for _, tt := range inputs {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = sanitizeOutput(tt.value)
			}
		})
	}
}

func BenchmarkSanitizeStderr(b *testing.B) {
	inputs := []struct {
		name  string
		value string
	}{
		{"short", "connection closed"},
		{"at_limit", strings.Repeat("x", 500)},
		{"over_limit", strings.Repeat("x", 1024)},
		{"empty", ""},
	}
	for _, tt := range inputs {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = sanitizeStderr(tt.value)
			}
		})
	}
}
