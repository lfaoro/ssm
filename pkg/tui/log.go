// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Log implements a Bubbletea component for displaying error and debug messages.
type Log struct {
	err          error
	debugLogs    []string
	debugActive  bool
	debugHistory int
	debugCount   int

	ErrStyle   lipgloss.Style
	DebugStyle lipgloss.Style
}

// DebugMsg is sent by AddLog to display a debug message.
type DebugMsg struct {
	Log string
}

// ErrorMsg is sent by AddError to display an error message.
type ErrorMsg struct {
	Err error
}

// LogOption configures a Log component.
type LogOption func(*Log)

// WithDebug enables or disables debug logging.
func WithDebug(debug bool) LogOption {
	return func(l *Log) {
		l.debugActive = debug
	}
}

// WithDebugHistory sets the number of debug messages retained.
func WithDebugHistory(length int) LogOption {
	return func(l *Log) {
		l.debugHistory = length
	}
}

// NewLog creates a new Log component with the given options.
func NewLog(opts ...LogOption) Log {
	l := Log{
		debugLogs:    make([]string, 0),
		err:          nil,
		debugActive:  false, // default
		debugHistory: 5,     // default
	}
	l.ErrStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1"))
	l.DebugStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	for _, opt := range opts {
		opt(&l)
	}
	return l
}

// AddLog sends a DebugMsg to the log component.
func AddLog(format string, args ...any) tea.Cmd {
	return func() tea.Msg {
		return DebugMsg{
			Log: fmt.Sprintf(format, args...),
		}
	}
}

// AddError sends an ErrorMsg to the log component.
func AddError(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Err: err}
	}
}

// ClearDebug sends messages to clear both debug and error state.
func ClearDebug() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ErrorMsg{Err: nil}
		},
		func() tea.Msg {
			return DebugMsg{Log: ""}
		},
	)
}

// ClearError sends a message to clear the current error state.
func ClearError() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ErrorMsg{Err: nil}
		},
	)
}

// Init initialises the log component.
func (l Log) Init() tea.Cmd {
	return AddLog("log: debug activated")
}

// Update handles DebugMsg and ErrorMsg for the log component.
func (l Log) Update(msg tea.Msg) (Log, tea.Cmd) {
	switch msg := msg.(type) {
	case DebugMsg:
		if !l.debugActive {
			return l, nil
		}
		l.debugCount++
		msgLog := fmt.Sprintf("%d: %s", l.debugCount, msg.Log)
		l.debugLogs = append(l.debugLogs, msgLog)
		if len(l.debugLogs) > l.debugHistory {
			l.debugLogs = l.debugLogs[len(l.debugLogs)-l.debugHistory:]
		}
	case ErrorMsg:
		l.err = msg.Err
	}
	return l, nil
}

// View renders the log component.
func (l Log) View() string {
	errMsg := func() string {
		if l.err != nil {
			var msg = l.err.Error()
			return l.ErrStyle.Render(msg)
		}
		return ""
	}
	debugMsg := func() string {
		if !l.debugActive {
			return ""
		}
		out := ""
		for i, log := range l.debugLogs {
			if len(l.debugLogs)-1 == i {
				// if last log, don't add a newline
				out += log
			} else {
				out += fmt.Sprintf("%s\n", log)
			}
		}
		out = l.DebugStyle.Render(out)
		return out
	}
	var out string
	if l.debugActive {
		out = errMsg() + "\n" + debugMsg()
	} else {
		out = errMsg()
	}
	out = strings.TrimSpace(out)
	return lipgloss.NewStyle().
		Padding(0, 0, 0, 1).
		Border(lipgloss.HiddenBorder(), true).
		BorderForeground(lipgloss.Color("240")).
		Render(out)
}
