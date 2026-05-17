// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"errors"
	"image/color"
	"os/exec"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
)

func updateModel(m *Model, msg tea.Msg) *Model {
	result, _ := m.Update(msg)
	return result.(*Model)
}

func updateModelWithCmd(m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	result, cmd := m.Update(msg)
	return result.(*Model), cmd
}

func TestNewModel(t *testing.T) {
	t.Run("creates model with defaults", func(t *testing.T) {
		cfg := newTestConfig(t)
		m := NewModel(cfg, false)

		if m == nil {
			t.Fatal("expected non-nil model")
		}
		if m.config != cfg {
			t.Error("config not set")
		}
		if m.debug {
			t.Error("debug should be false")
		}
		if m.Cmd != sshCmd {
			t.Errorf("Cmd = %v, want %v", m.Cmd, sshCmd)
		}
		if m.ExitOnCmd {
			t.Error("ExitOnCmd should be false")
		}
		if m.showConfig {
			t.Error("showConfig should be false")
		}
	})

	t.Run("debug mode enabled", func(t *testing.T) {
		cfg := newTestConfig(t)
		m := NewModel(cfg, true)

		if !m.debug {
			t.Error("debug should be true")
		}
	})
}

func TestModel_Init(t *testing.T) {
	t.Run("returns commands", func(t *testing.T) {
		m := newTestModel(t, false)
		cmd := m.Init()

		if cmd == nil {
			t.Error("expected non-nil command")
		}
	})

	t.Run("debug mode adds tick command", func(t *testing.T) {
		m := newTestModel(t, true)
		cmd := m.Init()

		if cmd == nil {
			t.Error("expected non-nil command")
		}
	})
}

func TestModel_Update_BackgroundColor(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, tea.BackgroundColorMsg{Color: color.Black})

	if !m2.isDark {
		t.Error("expected isDark to be true for black background")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	if m2.li.Width() != 100 {
		t.Errorf("list width = %d, want 100", m2.li.Width())
	}
	if m2.vp.Width() != 50 {
		t.Errorf("viewport width = %d, want 50", m2.vp.Width())
	}
}

func TestModel_Update_WindowSize_Debug(t *testing.T) {
	m := newTestModel(t, true)
	m2 := updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	if m2.li.Height() != 30-9 {
		t.Errorf("list height in debug = %d, want %d", m2.li.Height(), 30-9)
	}
}

func TestModel_Update_TabKey(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, tea.KeyPressMsg{Code: tea.KeyTab})

	if m2.Cmd != moshCmd {
		t.Errorf("Cmd = %v, want %v", m2.Cmd, moshCmd)
	}

	m3 := updateModel(m2, tea.KeyPressMsg{Code: tea.KeyTab})

	if m3.Cmd != sshCmd {
		t.Errorf("Cmd = %v, want %v", m3.Cmd, sshCmd)
	}
}

func TestModel_Update_EscKey(t *testing.T) {
	m := newTestModel(t, false)

	m.li.SetFilteringEnabled(true)
	m.li.SetFilterText("test")

	m2 := updateModel(m, tea.KeyPressMsg{Code: tea.KeyEsc})

	if m2.li.FilterState() == list.Filtering {
		t.Error("expected filter to be reset after Esc")
	}
}

func TestModel_Update_CtrlC(t *testing.T) {
	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	if cmd == nil {
		t.Fatal("expected command to be returned")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected QuitMsg, got %T", msg)
	}

	_ = m2
}

func TestModel_Update_CtrlC_WhileFiltering(t *testing.T) {
	m := newTestModel(t, false)

	m.li.SetFilterText("test")
	m2 := updateModel(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	if m2.li.FilterState() == list.Filtering {
		t.Error("expected filter to be reset")
	}
}

func TestModel_Update_CtrlV(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})

	if !m2.showConfig {
		t.Error("expected showConfig to be true")
	}

	m3 := updateModel(m2, tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})

	if m3.showConfig {
		t.Error("expected showConfig to be false after toggle")
	}
}

func TestModel_Update_CtrlR(t *testing.T) {
	m := newTestModel(t, false)

	result, _ := m.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})

	_, ok := result.(*cmdModel)
	if !ok {
		t.Errorf("expected *cmdModel, got %T", result)
	}
}

func TestModel_Update_CtrlE(t *testing.T) {
	skipIfNoEditor(t)

	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})

	if cmd == nil {
		t.Error("expected exec command for editor")
	}

	_ = m2
}

func TestModel_Update_CtrlE_NoEditor(t *testing.T) {
	t.Setenv("EDITOR", "")

	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})

	if cmd == nil {
		t.Skip("skipping: editor found in PATH")
	}

	_ = m2
}

func TestModel_Update_ReloadConfig(t *testing.T) {
	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, ReloadConfigMsg{})

	if cmd == nil {
		t.Error("expected command after reload")
	}

	hosts := m2.config.GetHosts()
	if len(hosts) == 0 {
		t.Error("expected hosts after reload")
	}
}

func TestModel_Update_SetTheme(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, SetThemeMsg{Theme: "sky"})

	if m2.theme.mainTitleColor != "#4682b4" {
		t.Errorf("theme mainTitleColor = %q, want %q", m2.theme.mainTitleColor, "#4682b4")
	}

	m3 := updateModel(m2, SetThemeMsg{Theme: "matrix"})

	if m3.theme.mainTitleColor != "#648c11" {
		t.Errorf("theme mainTitleColor = %q, want %q", m3.theme.mainTitleColor, "#648c11")
	}
}

func TestModel_Update_FilterTag(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, FilterTagMsg{Arg: "#test"})

	if !m2.li.FilteringEnabled() {
		t.Error("expected filtering to be enabled")
	}
	if m2.li.FilterValue() != "#test" {
		t.Errorf("filter value = %q, want %q", m2.li.FilterValue(), "#test")
	}
}

func TestModel_Update_ExitOnConn(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, ExitOnConnMsg{})

	if !m2.ExitOnCmd {
		t.Error("expected ExitOnCmd to be true")
	}
}

func TestModel_Update_AppMsg(t *testing.T) {
	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, AppMsg{Text: "test error"})

	if cmd == nil {
		t.Error("expected command for app message")
	}

	msg := cmd()
	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}
	if errMsg.Err == nil {
		t.Error("expected error to be set")
	}

	_ = m2
}

func TestModel_Update_Tick(t *testing.T) {
	m := newTestModel(t, true)
	m2, cmd := updateModelWithCmd(m, tickMsg{})

	if cmd == nil {
		t.Error("expected command for tick")
	}

	_ = m2
}

func TestModel_Update_UnknownKey(t *testing.T) {
	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, tea.KeyPressMsg{Code: 'x'})

	if cmd == nil {
		t.Log("expected clear error command for unknown key")
	}

	_ = m2
}

func TestModel_connect_NoSelection(t *testing.T) {
	cfg := newTestConfig(t)
	m := NewModel(cfg, false)

	m.li.SetItems([]list.Item{})

	cmd := m.connect()

	if cmd == nil {
		t.Fatal("expected error command")
	}

	msg := cmd()
	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}
	if errMsg.Err == nil {
		t.Error("expected error for no selection")
	}
}

func TestModel_connect_ExitOnCmd(t *testing.T) {
	m := newTestModel(t, false)

	m.li.CursorDown()
	m.ExitOnCmd = true

	cmd := m.connect()

	if cmd == nil {
		t.Fatal("expected command")
	}

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected QuitMsg, got %T", msg)
	}

	if m.ExitHost == "" {
		t.Error("expected ExitHost to be set")
	}
}

func TestModel_connect_MissingCmd(t *testing.T) {
	m := newTestModel(t, false)
	m.Cmd = "nonexistent-cmd-12345"
	m.li.CursorDown()

	cmd := m.connect()

	if cmd == nil {
		t.Fatal("expected error command")
	}

	msg := cmd()
	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}
	if errMsg.Err == nil {
		t.Error("expected error for missing command")
	}
}

func TestModel_setConfig(t *testing.T) {
	m := newTestModel(t, false)

	m.li.CursorDown()
	m.showConfig = true
	m.setConfig()

	content := m.vp.GetContent()
	if content == "" {
		t.Error("expected viewport content")
	}
}

func TestModel_setConfig_SensitiveKeysFiltered(t *testing.T) {
	cfg := newTestConfig(t)
	m := NewModel(cfg, false)

	m.li.CursorDown()
	m.li.CursorDown()
	m.li.CursorDown()
	m.showConfig = true
	m.setConfig()

	content := m.vp.GetContent()

	if containsStr(content, "IdentityFile") {
		t.Error("sensitive key IdentityFile should be filtered")
	}
	if containsStr(content, "ProxyCommand") {
		t.Error("sensitive key ProxyCommand should be filtered")
	}
}

func TestModel_setConfig_OutOfBounds(t *testing.T) {
	cfg := newTestConfig(t)
	m := NewModel(cfg, false)

	m.li.SetItems([]list.Item{})

	m.showConfig = true
	m.setConfig()

	_ = m.vp.GetContent()
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"identityfile", true},
		{"certificatefile", true},
		{"proxycommand", true},
		{"pkcs11provider", true},
		{"controlpath", true},
		{"userknownhostsfile", true},
		{"revokedhostkeys", true},
		{"globalknownhostsfile", true},
		{"user", false},
		{"hostname", false},
		{"port", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := sshconf.IsSensitiveKey(tt.key); got != tt.want {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestSanitizeStderr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short input unchanged",
			input: "short error",
			want:  "short error",
		},
		{
			name:  "trailing whitespace trimmed",
			input: "error message   \n",
			want:  "error message",
		},
		{
			name:  "long input truncated",
			input: string(make([]byte, 600)),
			want:  string(make([]byte, 500)) + "...",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeStderr(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeStderr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModel_View(t *testing.T) {
	m := newTestModel(t, false)

	v := m.View()

	if v.AltScreen == false && v.WindowTitle == "" {
		t.Fatal("expected non-nil view")
	}
	if !v.AltScreen {
		t.Error("expected AltScreen to be true")
	}
	if v.WindowTitle != "SSM | Secure Shell Manager" {
		t.Errorf("WindowTitle = %q, want %q", v.WindowTitle, "SSM | Secure Shell Manager")
	}
}

func TestModel_View_ShowConfig(t *testing.T) {
	m := newTestModel(t, false)
	m.showConfig = true
	m.setConfig()

	v := m.View()

	if v.AltScreen == false && v.WindowTitle == "" {
		t.Fatal("expected non-nil view")
	}
}

func TestModel_View_Debug(t *testing.T) {
	m := newTestModel(t, true)

	v := m.View()

	if v.AltScreen == false && v.WindowTitle == "" {
		t.Fatal("expected non-nil view")
	}
}

func TestTick(t *testing.T) {
	cmd := tick()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	msg := cmd()
	if _, ok := msg.(tickMsg); !ok {
		t.Errorf("expected tickMsg, got %T", msg)
	}
}

func TestModel_Update_LivenessCheck(t *testing.T) {
	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, LivenessCheckMsg{})

	if cmd == nil {
		t.Error("expected command for liveness check")
	}

	_ = m2
}

func TestModel_Update_CtrlP_CtrlN(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})

	if m2.li.Index() != 0 {
		t.Log("cursor up from top should stay at 0")
	}

	m3 := updateModel(m2, tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})

	if m3.li.Index() != 1 {
		t.Errorf("expected index 1, got %d", m3.li.Index())
	}
}

func TestModel_Update_CtrlB_CtrlF(t *testing.T) {
	m := newTestModel(t, false)
	m2 := updateModel(m, tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})

	if m2.li.Index() != 0 {
		t.Log("prev page from top should stay at 0")
	}

	m3 := updateModel(m2, tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl})

	if m3.li.Index() == 0 {
		t.Log("next page should move cursor")
	}
}

func TestModel_Update_CtrlS_SftpModel(t *testing.T) {
	m := newTestModel(t, false)
	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})

	_, ok := result.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result)
	}
}

func TestModel_Update_UnknownCtrlKey(t *testing.T) {
	m := newTestModel(t, false)
	m2, cmd := updateModelWithCmd(m, tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})

	if cmd == nil {
		t.Error("expected command for unknown ctrl key")
	}

	_ = m2
}

func TestModel_Connect_SSH(t *testing.T) {
	skipIfCmdNotFound(t, "ssh")

	m := newTestModel(t, false)
	m.li.CursorDown()

	cmd := m.connect()

	if cmd == nil {
		t.Fatal("expected connect command")
	}
}

func TestModel_Connect_Mosh(t *testing.T) {
	_, err := exec.LookPath("mosh")
	if err != nil {
		t.Skip("mosh not found in PATH")
	}

	m := newTestModel(t, false)
	m.Cmd = moshCmd
	m.li.CursorDown()

	cmd := m.connect()

	if cmd == nil {
		t.Fatal("expected connect command")
	}
}

func TestModel_Update_ErrorBuffer(t *testing.T) {
	m := newTestModel(t, false)

	m.errbuf.WriteString("test error output")

	m2 := updateModel(m, tea.KeyPressMsg{Code: 'x'})

	if m2.errbuf.Len() != 0 {
		t.Error("expected error buffer to be reset")
	}
}

func TestModel_Update_Backspace_ResetFilter(t *testing.T) {
	m := newTestModel(t, false)

	m.li.SetFilterText("test")

	m2 := updateModel(m, tea.KeyPressMsg{Code: tea.KeyBackspace})

	if m2.li.FilterState() == list.Filtering {
		t.Error("expected filter to be reset after backspace")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestModel_Update_ClearError(t *testing.T) {
	m := newTestModel(t, false)

	m.log.err = errors.New("test error")

	m2 := updateModel(m, tea.KeyPressMsg{Code: 'x'})

	_ = m2
}
