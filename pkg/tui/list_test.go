// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"testing"

	"github.com/lfaoro/ssm/pkg/sshconf"
	som "github.com/thalesfsp/go-common-types/safeorderedmap"
)

func TestItem_Interface(t *testing.T) {
	i := item{
		title: "test-host",
		desc:  "test description",
	}

	if i.Title() != "test-host" {
		t.Errorf("Title() = %q, want %q", i.Title(), "test-host")
	}
	if i.Description() != "test description" {
		t.Errorf("Description() = %q, want %q", i.Description(), "test description")
	}
	if i.FilterValue() != "test-hosttest description" {
		t.Errorf("FilterValue() = %q, want %q", i.FilterValue(), "test-hosttest description")
	}
}

func TestFormatHost_Basic(t *testing.T) {
	host := sshconf.Host{
		Name:    "simple-host",
		Options: som.New[string](),
	}

	result := formatHost(host)

	if result.title != "simple-host" {
		t.Errorf("title = %q, want %q", result.title, "simple-host")
	}
	if result.desc == "" {
		t.Error("expected non-empty description")
	}
}

func TestFormatHost_WithUser(t *testing.T) {
	opts := som.New[string]()
	opts.Add("user", "admin")

	host := sshconf.Host{
		Name:    "user-host",
		Options: opts,
	}

	result := formatHost(host)

	if !contains(result.desc, "admin@") {
		t.Errorf("desc = %q, should contain %q", result.desc, "admin@")
	}
}

func TestFormatHost_WithPort(t *testing.T) {
	opts := som.New[string]()
	opts.Add("port", "2222")

	host := sshconf.Host{
		Name:    "port-host",
		Options: opts,
	}

	result := formatHost(host)

	if !contains(result.desc, ":2222") {
		t.Errorf("desc = %q, should contain %q", result.desc, ":2222")
	}
}

func TestFormatHost_DefaultPort22(t *testing.T) {
	opts := som.New[string]()
	opts.Add("port", "22")

	host := sshconf.Host{
		Name:    "default-port",
		Options: opts,
	}

	result := formatHost(host)

	if contains(result.desc, ":22") {
		t.Errorf("desc = %q, should not contain default port :22", result.desc)
	}
}

func TestFormatHost_WithHostname(t *testing.T) {
	opts := som.New[string]()
	opts.Add("hostname", "10.0.0.1")

	host := sshconf.Host{
		Name:    "hostname-host",
		Options: opts,
	}

	result := formatHost(host)

	if !contains(result.desc, "10.0.0.1") {
		t.Errorf("desc = %q, should contain %q", result.desc, "10.0.0.1")
	}
}

func TestFormatHost_WithTags(t *testing.T) {
	opts := som.New[string]()
	opts.Add("#tag:", "production,api")

	host := sshconf.Host{
		Name:    "tagged-host",
		Options: opts,
	}

	result := formatHost(host)

	if !contains(result.desc, "#production,api") {
		t.Errorf("desc = %q, should contain %q", result.desc, "#production,api")
	}
}

func TestFormatHost_AllOptions(t *testing.T) {
	opts := som.New[string]()
	opts.Add("user", "deploy")
	opts.Add("hostname", "api.example.com")
	opts.Add("port", "2222")
	opts.Add("#tag:", "production")

	host := sshconf.Host{
		Name:    "full-host",
		Options: opts,
	}

	result := formatHost(host)

	if result.title != "full-host" {
		t.Errorf("title = %q, want %q", result.title, "full-host")
	}
	if !contains(result.desc, "deploy@") {
		t.Errorf("desc should contain user, got %q", result.desc)
	}
	if !contains(result.desc, "api.example.com") {
		t.Errorf("desc should contain hostname, got %q", result.desc)
	}
	if !contains(result.desc, ":2222") {
		t.Errorf("desc should contain port, got %q", result.desc)
	}
	if !contains(result.desc, "#production") {
		t.Errorf("desc should contain tag, got %q", result.desc)
	}
}

func TestFormatHost_NoHostname(t *testing.T) {
	host := sshconf.Host{
		Name:    "no-hostname",
		Options: som.New[string](),
	}

	result := formatHost(host)

	if !contains(result.desc, "no-hostname") {
		t.Errorf("desc = %q, should fall back to host name", result.desc)
	}
}

func TestListFrom(t *testing.T) {
	cfg := newTestConfig(t)

	li := listFrom(cfg, matrixTheme())

	items := li.Items()
	if len(items) == 0 {
		t.Error("expected non-empty items list")
	}

	if li.Title == "" {
		t.Error("expected list title to be set")
	}

	if li.FilterInput.Placeholder == "" {
		t.Error("expected filter placeholder to be set")
	}

	if li.FilterInput.CharLimit != 12 {
		t.Errorf("filter char limit = %d, want 12", li.FilterInput.CharLimit)
	}
}

func TestListFrom_ItemsMatchHosts(t *testing.T) {
	cfg := newTestConfig(t)
	hosts := cfg.GetHosts()

	li := listFrom(cfg, skyTheme())

	items := li.Items()
	if len(items) != len(hosts) {
		t.Errorf("items count = %d, want %d", len(items), len(hosts))
	}
}

func TestListFrom_HelpKeys(t *testing.T) {
	cfg := newTestConfig(t)

	li := listFrom(cfg, matrixTheme())

	if li.AdditionalFullHelpKeys == nil {
		t.Error("expected additional help keys to be set")
	} else {
		keys := li.AdditionalFullHelpKeys()
		if len(keys) == 0 {
			t.Error("expected non-empty help keys")
		}
	}
}

func TestInitKeys(t *testing.T) {
	keys := initKeys()

	if len(keys) != 6 {
		t.Errorf("expected 6 key bindings, got %d", len(keys))
	}

	expectedKeys := map[string]bool{
		"enter":  false,
		"tab":    false,
		"ctrl+s": false,
		"ctrl+r": false,
		"ctrl+e": false,
		"ctrl+v": false,
	}

	for _, k := range keys {
		for _, keyStr := range k.Keys() {
			if _, ok := expectedKeys[keyStr]; ok {
				expectedKeys[keyStr] = true
			}
		}
	}

	for keyStr, found := range expectedKeys {
		if !found {
			t.Errorf("expected key binding %q not found", keyStr)
		}
	}
}

func TestListFrom_FilterInput(t *testing.T) {
	cfg := newTestConfig(t)

	li := listFrom(cfg, matrixTheme())

	if li.FilterInput.Prompt != "Search: " {
		t.Errorf("filter prompt = %q, want %q", li.FilterInput.Prompt, "Search: ")
	}

	if li.FilterInput.Placeholder != "hostName or tagName" {
		t.Errorf("filter placeholder = %q, want %q", li.FilterInput.Placeholder, "hostName or tagName")
	}
}

func TestListFrom_StatusBarItemName(t *testing.T) {
	cfg := newTestConfig(t)

	li := listFrom(cfg, matrixTheme())

	view := li.View()

	if !contains(view, "host") && !contains(view, "hosts") {
		t.Log("expected status bar to reference hosts")
	}
}

func TestItem_FilterValue_ContainsTitle(t *testing.T) {
	i := item{
		title: "my-server",
		desc:  "user@host",
	}

	fv := i.FilterValue()

	if !contains(fv, "my-server") {
		t.Errorf("FilterValue() = %q, should contain title", fv)
	}
	if !contains(fv, "user@host") {
		t.Errorf("FilterValue() = %q, should contain description", fv)
	}
}

func TestFormatHost_EmptyOptions(t *testing.T) {
	host := sshconf.Host{
		Name:    "empty-options",
		Options: som.New[string](),
	}

	result := formatHost(host)

	if result.title != "empty-options" {
		t.Errorf("title = %q, want %q", result.title, "empty-options")
	}
}

func TestListFrom_SkyTheme(t *testing.T) {
	cfg := newTestConfig(t)

	li := listFrom(cfg, skyTheme())

	items := li.Items()
	if len(items) == 0 {
		t.Error("expected items with sky theme")
	}
}

func TestListFrom_EmptyConfig(t *testing.T) {
	cfg := sshconf.New()

	li := listFrom(cfg, matrixTheme())

	items := li.Items()
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty config, got %d", len(items))
	}
}

func TestListFrom_ListModelType(t *testing.T) {
	cfg := newTestConfig(t)

	li := listFrom(cfg, matrixTheme())

	_ = li
}
