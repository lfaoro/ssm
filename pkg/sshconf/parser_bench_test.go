// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package sshconf_test

import (
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

func BenchmarkParsePath(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		cfg := sshconf.New()
		_ = cfg.ParsePath("../../data/config_example")
	}
}

func BenchmarkGetHost(b *testing.B) {
	cfg := sshconf.New()
	if err := cfg.ParsePath("../../data/config_example"); err != nil {
		b.Fatalf("ParsePath: %v", err)
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = cfg.GetHost("prod-api")
	}
}

func BenchmarkGetHosts(b *testing.B) {
	cfg := sshconf.New()
	if err := cfg.ParsePath("../../data/config_example"); err != nil {
		b.Fatalf("ParsePath: %v", err)
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = cfg.GetHosts()
	}
}

func BenchmarkGetParamFor(b *testing.B) {
	cfg := sshconf.New()
	if err := cfg.ParsePath("../../data/config_example"); err != nil {
		b.Fatalf("ParsePath: %v", err)
	}
	host := cfg.GetHost("prod-api")
	b.ReportAllocs()
	for b.Loop() {
		_ = cfg.GetParamFor(host, "port")
	}
}

func BenchmarkRemoveComments(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"no_comment", "HostName example.com"},
		{"inline_comment", "HostName example.com # comment"},
		{"only_comment", "#tag: production,api"},
		{"empty", ""},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sshconf.RemoveComments(tt.input)
			}
		})
	}
}

func BenchmarkIsSensitiveKey(b *testing.B) {
	keys := []string{
		"identityfile",
		"hostname",
		"proxycommand",
		"user",
		"certificatefile",
		"port",
	}
	for _, k := range keys {
		b.Run(k, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = sshconf.IsSensitiveKey(k)
			}
		})
	}
}
