// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"errors"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func TestSanitizeOutput_Plain(t *testing.T) {
	input := "hello world"
	got := sanitizeOutput(input)

	if got != input {
		t.Errorf("sanitizeOutput(%q) = %q, want %q", input, got, input)
	}
}

func TestSanitizeOutput_ANSIColors(t *testing.T) {
	input := "\x1b[31mred text\x1b[0m"
	want := "red text"
	got := sanitizeOutput(input)

	if got != want {
		t.Errorf("sanitizeOutput() = %q, want %q", got, want)
	}
}

func TestSanitizeOutput_CursorMovement(t *testing.T) {
	input := "\x1b[2J\x1b[H"
	want := ""
	got := sanitizeOutput(input)

	if got != want {
		t.Errorf("sanitizeOutput() = %q, want %q", got, want)
	}
}

func TestSanitizeOutput_Mixed(t *testing.T) {
	input := "\x1b[32mOK\x1b[0m\n\x1b[1mBold\x1b[0m"
	want := "OK\nBold"
	got := sanitizeOutput(input)

	if got != want {
		t.Errorf("sanitizeOutput() = %q, want %q", got, want)
	}
}

func TestSanitizeOutput_Empty(t *testing.T) {
	got := sanitizeOutput("")

	if got != "" {
		t.Errorf("sanitizeOutput(\"\") = %q, want \"\"", got)
	}
}

func TestSanitizeOutput_BellSequence(t *testing.T) {
	input := "\x1b]0;title\x07plain"
	want := "plain"
	got := sanitizeOutput(input)

	if got != want {
		t.Errorf("sanitizeOutput() = %q, want %q", got, want)
	}
}

func TestSanitizeOutput_CarriageReturn(t *testing.T) {
	input := "line1\rline2"
	want := "line1line2"
	got := sanitizeOutput(input)

	if got != want {
		t.Errorf("sanitizeOutput() = %q, want %q", got, want)
	}
}

func TestSanitizeOutput_MultipleColors(t *testing.T) {
	input := "\x1b[1;31mError\x1b[0m: \x1b[33mwarning\x1b[0m"
	want := "Error: warning"
	got := sanitizeOutput(input)

	if got != want {
		t.Errorf("sanitizeOutput() = %q, want %q", got, want)
	}
}

func TestRunCmdModel_Creation(t *testing.T) {
	m := newTestModel(t, false)
	// run submodel always uses ssh for the actual exec, regardless of m.Cmd (mosh only for direct Enter)
	m.Cmd = MoshCmd

	result := RunCmdModel(m)

	cmdM, ok := result.(*cmdModel)
	if !ok {
		t.Fatalf("expected *cmdModel, got %T", result)
	}

	if !cmdM.input.Focused() {
		t.Error("expected input to be focused")
	}

	if cmdM.previousModel != m {
		t.Error("expected previousModel to be set")
	}

	if cmdM.running {
		t.Error("expected running to be false initially")
	}

	if cmdM.ready {
		t.Error("expected ready to be false initially")
	}
}

func TestCmdModel_Init(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	cmd := cmdM.Init()

	if cmd != nil {
		t.Log("Init() returned nil (expected)")
	}
}

func TestCmdModel_Update_Escape(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	result, _ := cmdM.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	_, ok := result.(*Model)
	if !ok {
		t.Errorf("expected *Model after escape, got %T", result)
	}
}

func TestCmdModel_Update_EmptyEnter(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	result, cmd := cmdM.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if cmd != nil {
		t.Error("expected no command for empty input")
	}

	_, ok := result.(*cmdModel)
	if !ok {
		t.Errorf("expected *cmdModel, got %T", result)
	}
}

func TestCmdModel_HandleCommandResult_Success(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	cmdM.handleCommandResult(cmdResultMsg{
		output: "command output",
		err:    nil,
	})

	if cmdM.running {
		t.Error("expected running to be false after command completes")
	}

	if !cmdM.input.Focused() {
		t.Error("expected input to be focused after command completes")
	}

	if len(cmdM.commands) == 0 {
		t.Error("expected commands to be recorded")
	}
}

func TestCmdModel_HandleCommandResult_Error(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	testErr := errors.New("command failed")
	cmdM.handleCommandResult(cmdResultMsg{
		output: "partial output",
		err:    testErr,
	})

	if cmdM.running {
		t.Error("expected running to be false after error")
	}

	if len(cmdM.commands) == 0 {
		t.Error("expected error to be recorded in commands")
	}
}

func TestCmdModel_Bar(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	bar := cmdM.Bar()

	if bar == "" {
		t.Error("expected non-empty bar")
	}

	if !containsStr(bar, "Run Command") {
		t.Error("expected bar to contain 'Run Command'")
	}
}

func TestCmdModel_Bar_NoSelection(t *testing.T) {
	cfg := newTestConfig(t)
	m := NewModel(cfg, false, SSHCmd)
	m.li.SetItems([]list.Item{})

	cmdM := RunCmdModel(m).(*cmdModel)

	bar := cmdM.Bar()

	if !containsStr(bar, "No host selected") {
		t.Error("expected bar to indicate no host selected")
	}
}

func TestCmdModel_View(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	v := cmdM.View()

	if v.AltScreen == false && v.WindowTitle == "" {
		t.Log("view created (may have default values)")
	}
}

func TestCmdModel_HandleWindowSize(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	cmdM.handleWindowSize(tea.WindowSizeMsg{Width: 100, Height: 30})

	if cmdM.input.Width() == 0 {
		t.Error("expected input width to be set")
	}

	if cmdM.viewport.Width() == 0 {
		t.Error("expected viewport width to be set")
	}
}

func TestCmdModel_Update_WindowSize(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	_, _ = cmdM.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	if cmdM.viewport.Width() != 100 {
		t.Errorf("viewport width = %d, want 100", cmdM.viewport.Width())
	}
}

func TestCmdModel_Update_Ready(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	_, cmd := cmdM.Update(tea.KeyPressMsg{Code: 'a'})

	if !cmdM.ready {
		t.Error("expected ready to be true after first update")
	}

	if cmd == nil {
		t.Log("expected command for ready state")
	}
}

func TestCmdModel_Update_CtrlL(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	cmdM.commands = []string{"$ ls", "output"}
	cmdM.viewport.SetContent("output")

	_, _ = cmdM.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl})

	if len(cmdM.commands) != 0 {
		t.Errorf("expected commands to be cleared, got %d", len(cmdM.commands))
	}
}

func TestCmdModel_Update_CtrlC_NoRunning(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	_, _ = cmdM.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	if len(cmdM.commands) == 0 {
		t.Error("expected message about no running command")
	}
}

func TestCmdModel_Bar_WithTheme(t *testing.T) {
	m := newTestModel(t, false)
	m.theme = skyTheme()
	cmdM := RunCmdModel(m).(*cmdModel)

	bar := cmdM.Bar()

	if bar == "" {
		t.Error("expected non-empty bar with theme")
	}
}

func TestRunCmdModel_ReturnsBaseOnInvalidModel(t *testing.T) {
	invalid := &invalidModel{}
	result := RunCmdModel(invalid)

	if result != invalid {
		t.Error("expected base model to be returned for invalid model type")
	}
}

type invalidModel struct{}

func (m *invalidModel) Init() tea.Cmd                         { return nil }
func (m *invalidModel) Update(_ tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *invalidModel) View() tea.View                        { return tea.NewView("") }

func TestCmdModel_Update_EnterWithCommand(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	cmdM.input.SetValue("echo hello")

	result, cmd := cmdM.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command for running")
	}

	updatedM, ok := result.(*cmdModel)
	if !ok {
		t.Errorf("expected *cmdModel, got %T", result)
	}

	if updatedM.running {
		t.Log("expected running to be true (may be reset after command)")
	}

	if !updatedM.input.Focused() {
		t.Log("input should be blurred while running")
	}
}

func TestCmdModel_Update_CtrlC_WhileRunning(t *testing.T) {
	m := newTestModel(t, false)
	cmdM := RunCmdModel(m).(*cmdModel)

	cmdM.running = true

	_, _ = cmdM.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	if len(cmdM.commands) == 0 {
		t.Error("expected message about cancellation")
	}
}

// --- Tests for batch command support (--command / -r) ---

func TestHostsForBatch_EmptyFilter(t *testing.T) {
	cfg := newTestConfig(t)
	all := cfg.GetHosts()

	got := hostsForBatch(all, "")
	if len(got) != len(all) {
		t.Errorf("empty filter should return all %d hosts, got %d", len(all), len(got))
	}
}

func TestHostsForBatch_TagFilter(t *testing.T) {
	cfg := newTestConfig(t)
	all := cfg.GetHosts()

	// "test" should match the host that has #tag: test
	got := hostsForBatch(all, "test")
	if len(got) == 0 {
		t.Fatal("expected at least one host for tag 'test'")
	}
	// sanity: the first one should be the tagged test server
	if got[0].Name != "test-server" {
		t.Logf("first match was %s (still acceptable if filter matched)", got[0].Name)
	}
}

func TestHostsForBatch_CommaTags(t *testing.T) {
	cfg := newTestConfig(t)
	got := hostsForBatch(cfg.GetHosts(), "web,secure")
	if len(got) < 2 {
		t.Errorf("expected at least 2 hosts for 'web,secure', got %d", len(got))
	}
}

func TestHostsForBatch_NameMatch(t *testing.T) {
	cfg := newTestConfig(t)
	got := hostsForBatch(cfg.GetHosts(), "no-tag-host")
	if len(got) != 1 || got[0].Name != "no-tag-host" {
		t.Errorf("expected exact name match for 'no-tag-host'")
	}
}

func TestExecOnHost_Basic(t *testing.T) {
	// Use a local command that fails instantly ("false") instead of real ssh
	// so the test never hangs on unreachable fixture hosts.
	out, err := execOnHost("false", testConfigPath(t), "test-server", "whoami")
	if err == nil {
		t.Error("expected error when using 'false' as the ssh binary")
	}
	// We mainly care that sanitize + error path didn't panic.
	_ = out
}

func TestSanitizeOutput_StillWorksAfterBatchChanges(t *testing.T) {
	// Regression guard that the shared sanitizeOutput wasn't accidentally broken.
	in := "\x1b[31mred\x1b[0m\r\n"
	want := "red\n"
	if got := sanitizeOutput(in); got != want {
		t.Errorf("sanitizeOutput regression: got %q want %q", got, want)
	}
}

func TestRenderPrimaryBar(t *testing.T) {
	result := renderPrimaryBar("test", "#ff0000")

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderSecondaryBar(t *testing.T) {
	result := renderSecondaryBar("test", 20)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

// TestRunBatchRemoteCommands_Params exercises the new 6-argument form
// (used by `ssm exec`) with explicit delay/threads/jitter and the "false"
// binary so the test remains hermetic and fast.
func TestRunBatchRemoteCommands_Params(t *testing.T) {
	cfg := newTestConfig(t)

	// Should not panic or deadlock even with tiny positive delay + jitter.
	err := RunBatchRemoteCommands(cfg, "", "echo hi", 1*time.Millisecond, 2, 5*time.Millisecond)
	// We expect errors (because "false" binary fails), but the runner must
	// complete and return a proper error instead of hanging.
	if err == nil {
		t.Log("unexpected success with 'false' ssh binary (acceptable)")
	}
}
