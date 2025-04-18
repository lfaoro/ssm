package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type Log struct {
	err          error
	debugLogs    []string
	debugActive  bool
	debugHistory int
	debugCount   int

	ErrStyle   lipgloss.Style
	DebugStyle lipgloss.Style
}

type DebugMsg struct {
	Log string
}

type ErrorMsg struct {
	Err error
}

type LogOption func(*Log)

func WithDebug(debug bool) LogOption {
	return func(l *Log) {
		l.debugActive = debug
	}
}

func WithDebugHistory(length int) LogOption {
	return func(l *Log) {
		l.debugHistory = length
	}
}

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

func AddLog(format string, args ...any) tea.Cmd {
	return func() tea.Msg {
		return DebugMsg{
			Log: fmt.Sprintf(format, args...),
		}
	}
}

func AddError(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Err: err}
	}
}

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

func ClearError() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ErrorMsg{Err: nil}
		},
	)
}

func (l Log) Init() tea.Cmd {
	return AddLog("log: debug activated")
}

func (l Log) Update(msg tea.Msg) (Log, tea.Cmd) {
	switch msg := msg.(type) {
	case DebugMsg:
		l.debugCount++
		msgLog := fmt.Sprintf("%d: %s", l.debugCount, msg.Log)
		if l.debugActive {
			l.debugLogs = append(l.debugLogs, msgLog)
		}
		if len(l.debugLogs) > l.debugHistory {
			l.debugLogs = l.debugLogs[len(l.debugLogs)-l.debugHistory:]
		}
	case ErrorMsg:
		l.err = msg.Err
	}
	return l, nil
}

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
	out := errMsg() + "\n" + debugMsg()
	return lipgloss.NewStyle().
		Padding(0).
		Border(lipgloss.NormalBorder(), l.debugActive).
		BorderForeground(lipgloss.Color("240")).
		Render(out)
}
