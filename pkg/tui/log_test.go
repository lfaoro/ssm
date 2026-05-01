// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"errors"
	"testing"
)

func TestLog_DebugMessages(t *testing.T) {
	t.Run("debug active stores messages", func(t *testing.T) {
		l := NewLog(WithDebug(true), WithDebugHistory(3))

		l, _ = l.Update(DebugMsg{Log: "message one"})
		l, _ = l.Update(DebugMsg{Log: "message two"})

		if len(l.debugLogs) != 2 {
			t.Fatalf("expected 2 debug logs, got %d", len(l.debugLogs))
		}
	})

	t.Run("debug inactive ignores messages", func(t *testing.T) {
		l := NewLog()

		l, _ = l.Update(DebugMsg{Log: "should be ignored"})
		if len(l.debugLogs) != 0 {
			t.Fatalf("expected 0 debug logs, got %d", len(l.debugLogs))
		}
	})

	t.Run("ring buffer truncates old messages", func(t *testing.T) {
		l := NewLog(WithDebug(true), WithDebugHistory(2))

		l, _ = l.Update(DebugMsg{Log: "first"})
		l, _ = l.Update(DebugMsg{Log: "second"})
		l, _ = l.Update(DebugMsg{Log: "third"})

		if len(l.debugLogs) != 2 {
			t.Fatalf("expected 2 debug logs, got %d", len(l.debugLogs))
		}

		// verify the logs directly rather than via View (which adds ANSI escapes)
		foundSecond := false
		foundThird := false
		for _, log := range l.debugLogs {
			if contains(log, "second") {
				foundSecond = true
			}
			if contains(log, "third") {
				foundThird = true
			}
		}
		if !foundSecond || !foundThird {
			t.Errorf("expected ring buffer to contain second and third, got %v", l.debugLogs)
		}
	})

	t.Run("debugCount increments", func(t *testing.T) {
		l := NewLog(WithDebug(true))

		l, _ = l.Update(DebugMsg{Log: "msg1"})
		l, _ = l.Update(DebugMsg{Log: "msg2"})

		if l.debugCount != 2 {
			t.Errorf("debugCount = %d, want 2", l.debugCount)
		}
	})
}

func TestLog_ErrorMessages(t *testing.T) {
	t.Run("sets error", func(t *testing.T) {
		l := NewLog()

		testErr := errors.New("test error")
		l, _ = l.Update(ErrorMsg{Err: testErr})

		if l.err == nil {
			t.Fatal("expected error to be set")
		}
		if !errors.Is(l.err, testErr) {
			t.Errorf("err = %v, want %v", l.err, testErr)
		}
	})

	t.Run("clears error", func(t *testing.T) {
		l := NewLog()

		l, _ = l.Update(ErrorMsg{Err: errors.New("test error")})
		l, _ = l.Update(ErrorMsg{Err: nil})

		if l.err != nil {
			t.Errorf("expected error to be nil, got %v", l.err)
		}
	})

	t.Run("view with error shows message", func(t *testing.T) {
		l := NewLog()

		l, _ = l.Update(ErrorMsg{Err: errors.New("something broke")})
		view := l.View()

		if !contains(view, "something broke") {
			t.Error("expected view to contain error message")
		}
	})

	t.Run("view without error is empty", func(t *testing.T) {
		l := NewLog()
		view := l.View()
		if contains(view, "error") || contains(view, "Error") {
			t.Errorf("expected empty view, got %q", view)
		}
	})
}

func TestAddError(t *testing.T) {
	t.Run("returns ErrorMsg", func(t *testing.T) {
		testErr := errors.New("test")
		cmd := AddError(testErr)
		msg := cmd()

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}
		if !errors.Is(errMsg.Err, testErr) {
			t.Errorf("err = %v, want %v", errMsg.Err, testErr)
		}
	})
}

func TestAddLog(t *testing.T) {
	t.Run("returns DebugMsg with formatted message", func(t *testing.T) {
		cmd := AddLog("hello %s", "world")
		msg := cmd()

		debugMsg, ok := msg.(DebugMsg)
		if !ok {
			t.Fatalf("expected DebugMsg, got %T", msg)
		}
		if debugMsg.Log != "hello world" {
			t.Errorf("log = %q, want %q", debugMsg.Log, "hello world")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewLog_Defaults(t *testing.T) {
	l := NewLog()

	if l.debugActive {
		t.Error("expected debugActive to be false by default")
	}
	if l.debugHistory != 5 {
		t.Errorf("debugHistory = %d, want 5", l.debugHistory)
	}
}

func TestClearError(t *testing.T) {
	t.Run("clears error in log", func(t *testing.T) {
		l := NewLog()
		l, _ = l.Update(ErrorMsg{Err: errors.New("test error")})

		if l.err == nil {
			t.Fatal("error should be set before clear")
		}

		cmd := ClearError()
		cmd()

		l, _ = l.Update(ErrorMsg{Err: nil})

		if l.err != nil {
			t.Errorf("expected error to be cleared, got %v", l.err)
		}
	})
}
