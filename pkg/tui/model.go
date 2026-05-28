// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

// Package tui defines the terminal user interface of this application.
package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/lfaoro/ssm/pkg/sshconf"
)

// Model is the main Bubbletea application model.
type Model struct {
	config     *sshconf.Config
	showConfig bool
	theme      theme

	li list.Model
	vp viewport.Model

	Cmd       SysCmd
	ExitOnCmd bool
	ExitHost  string

	debug bool
	log   Log

	pingResults map[string]string

	errbuf bytes.Buffer
	isDark bool
}

// NewModel creates a new application Model.
func NewModel(config *sshconf.Config, debug bool) *Model {
	m := &Model{}
	m.debug = debug
	m.config = config
	m.theme = skyTheme()
	m.li = listFrom(m.config, m.theme)
	m.log = NewLog(WithDebug(debug))
	m.Cmd = sshCmd // defaults to ssh
	m.pingResults = make(map[string]string)
	m.vp = viewport.New()
	m.vp.SetWidth(40)
	m.vp.SetHeight(20)
	return m
}

// Init initialises the model and returns the initial commands.
func (m *Model) Init() tea.Cmd {
	m.syncViewportStyle()
	cmds := []tea.Cmd{
		tea.RequestCapability("keyboard_enhancements"),
	}
	m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.Cmd))
	if m.debug {
		cmds = append(cmds, tick())
	}
	return tea.Batch(cmds...)
}

// Update handles all messages and returns the updated model.
//
//nolint:gocyclo // TEA message dispatchers are naturally complex (large switch); threshold is 55 and this change only added one guarded case.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()
		if m.debug {
			cmds = append(cmds, AddLog("debug: isdarkbg %v", m.isDark))
		}
	case tea.WindowSizeMsg:
		var errSize = 1
		if m.log.err != nil {
			errSize = 3
		}
		m.li.SetSize(msg.Width, msg.Height-errSize)
		if m.debug {
			m.li.SetSize(msg.Width, msg.Height-9)
		}

		m.vp.SetHeight(m.li.Height())
		m.vp.SetWidth(msg.Width / 2)
		m.syncViewportStyle()

		if m.log.err != nil {
			cmds = append(cmds, tea.RequestWindowSize)
		}
	case tickMsg:
		return m, tea.Batch(tick(),
			AddLog("ticking..."))
	case AppMsg:
		return m, AddError(fmt.Errorf("%s", msg.Text))
	case LivenessCheckMsg:
		return m, pingAllCmd(m)
	case ExitOnConnMsg:
		m.ExitOnCmd = true
		return m, AddLog("exit true")
	case FilterTagMsg:
		m.li.SetFilterText(msg.Arg)
		m.li.SetFilteringEnabled(true)
		return m, AddLog("filter true")
	case ReloadConfigMsg:
		err := m.config.ParsePath(m.config.GetPath())
		if err != nil {
			return m, AddError(err)
		}
		m.li = listFrom(m.config, m.theme)
		m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.Cmd))
		return m, AddLog("reloading config")
	case ShowConfigMsg:
		m.showConfig = true
		return m, nil
	case SetThemeMsg:
		m.theme = themes[msg.Theme]
		m.li = listFrom(m.config, m.theme)
		return m, nil
	case PingResultMsg:
		m.pingResults[msg.Host] = msg.Latency
		refreshList(m)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyTab:
			if m.Cmd == sshCmd {
				m.Cmd = moshCmd
				m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.Cmd))
			} else {
				m.Cmd = sshCmd
				m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.Cmd))
			}
		case tea.KeyEnter:
			if m.li.FilterState() == list.Filtering {
				if m.li.FilterValue() == "" {
					m.li.ResetFilter()
				}
				break
			}
			conncmd := m.connect()
			return m, tea.Batch(
				conncmd,
				AddError(fmt.Errorf("%s", m.errbuf.String())),
			)
		case tea.KeyEsc, tea.KeyBackspace:
			if m.li.FilteringEnabled() {
				m.li.ResetFilter()
				return m, nil
			}
		case 'q':
			if m.li.FilterState() != list.Filtering {
				return m, tea.Quit
			}
		case 'p':
			if m.li.FilterState() != list.Filtering {
				if msg.Mod == tea.ModShift {
					return m, pingAllCmd(m)
				}
				if msg.Mod == 0 {
					return m, pingSelectedCmd(m)
				}
			}
		case 'P':
			if m.li.FilterState() != list.Filtering {
				return m, pingAllCmd(m)
			}
		case 'y', 'Y':
			if m.li.FilterState() != list.Filtering {
				return m, m.copySelected()
			}
		}
		switch msg.Mod {
		// we're only interested in ctrl+<key>
		case tea.ModCtrl:
			switch msg.Code {
			case 'c':
				if m.li.FilterState() == list.Filtering ||
					m.li.IsFiltered() {
					m.li.ResetFilter()
					return m, nil
				}
				return m, tea.Quit

			// emacs keybinds
			case 'p':
				m.li.CursorUp()
			case 'n':
				m.li.CursorDown()
			case 'b':
				m.li.PrevPage()
			case 'f':
				m.li.NextPage()

			case 'e':
				confFile := m.config.GetPath()
				editorPath := os.Getenv("EDITOR")
				if editorPath != "" {
					if path, err := exec.LookPath(editorPath); err == nil {
						editorPath = path
					} else {
						editorPath = ""
					}
				}
				if editorPath == "" {
					for _, cmd := range []string{"vim", "vi", "nano", "ed"} {
						if path, err := exec.LookPath(cmd); err == nil {
							editorPath = path
							break
						}
					}
				}
				if editorPath == "" {
					return m, AddError(errors.New("env EDITOR not set, nor any editor found in PATH"))
				}
				cmd := exec.CommandContext(context.Background(), editorPath, confFile) //nolint:gosec
				cmd.Dir = filepath.Dir(confFile)
				cmd.Stderr = &m.errbuf
				execCmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
					logCmd := AddLog("%v", err)
					var errCmd tea.Cmd
					if err != nil {
						errCmd = AddError(err)
					}
					return tea.Batch(logCmd, errCmd)
				})
				return m, tea.Sequence(
					execCmd,
					func() tea.Msg {
						return ReloadConfigMsg{}
					},
				)
			case 'r':
				return RunCmdModel(m), nil
			case 's':
				return SftpModel(m), nil
			case 'v':
				m.showConfig = !m.showConfig
				m.syncViewportStyle()
			case 'y':
				return SyncModel(m), nil
			default:
				return m, AddError(fmt.Errorf("that's an interesting key combo! %s", msg))
			}
		default:
			cmds = append(cmds, ClearError())
		}
	}
	if len(m.errbuf.Bytes()) > 0 {
		cmds = append(cmds,
			AddError(fmt.Errorf("%v", m.errbuf.String())),
		)
		m.errbuf.Reset()
	}

	m.li, cmd = m.li.Update(msg)
	cmds = append(cmds, cmd)

	if m.showConfig {
		m.setConfig()
	}
	m.vp, cmd = m.vp.Update(msg)
	cmds = append(cmds, cmd)

	m.log, cmd = m.log.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) connect() tea.Cmd {
	host, ok := m.li.SelectedItem().(item)
	if !ok {
		return AddError(errors.New("unable to find selected item: open bug report"))
	}
	if m.ExitOnCmd {
		m.ExitHost = strings.TrimSpace(host.title)
		return tea.Quit
	}

	cmdPath, err := exec.LookPath(m.Cmd.String())
	if err != nil {
		return AddError(fmt.Errorf("can't find `%s` cmd in your path: %w", m.Cmd, err))
	}

	var cmd *exec.Cmd
	cmd = exec.CommandContext(context.Background(), cmdPath, "-F", m.config.GetPath(), "--", host.title) //nolint:gosec
	if m.Cmd == moshCmd {
		cmd = exec.CommandContext(context.Background(), //nolint:gosec
			cmdPath,
			"--",
			host.title,
		)
		cmd.Env = append(os.Environ(), "SSH_CONFIG="+m.config.GetPath())
	}

	cmd.Stderr = &m.errbuf
	execmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
		msg := fmt.Sprintf("connection closed: %v", host.title)
		if err != nil {
			msg += fmt.Sprintf(", err: %v", err)
		}
		if sanitized := sanitizeStderr(m.errbuf.String()); sanitized != "" {
			msg += "\n" + sanitized
		}
		return ErrorMsg{Err: fmt.Errorf("%s", msg)}
	})
	return execmd
}

// copySelected copies the currently selected host's name (the SSH config "Host" alias)
// to the system clipboard using atotto/clipboard. On success it shows a transient
// status message. Clipboard unavailability (common in headless/CI/SSH sessions) is
// handled gracefully with a status message instead of a hard error.
func (m *Model) copySelected() tea.Cmd {
	host, ok := m.li.SelectedItem().(item)
	if !ok {
		return AddError(errors.New("no host selected"))
	}
	if err := clipboard.WriteAll(host.title); err != nil {
		// Graceful degradation for headless environments (no xclip/wl-clipboard etc.)
		m.li.NewStatusMessage("Clipboard unavailable in this environment")
		if m.debug {
			return AddLog("clipboard write failed: %v", err)
		}
		return nil
	}
	m.li.NewStatusMessage("Copied: " + host.title)
	if m.debug {
		return AddLog("copied host: %s", host.title)
	}
	return nil
}

func (m *Model) setConfig() {
	i := m.li.GlobalIndex()
	hosts := m.config.GetHosts()
	if i < 0 || i >= len(hosts) {
		return
	}
	host := hosts[i]
	var out strings.Builder
	keyStyle := lg.NewStyle().
		Foreground(lg.Color("#4682b4"))
	for i, k := range host.Options.Keys() {
		if sshconf.IsSensitiveKey(k) {
			continue
		}
		out.WriteString(keyStyle.Render(k))
		out.WriteByte(' ')
		out.WriteString(host.Options.Values()[i])
		out.WriteByte('\n')
	}
	m.vp.SetContent(out.String())
}

func (m *Model) syncViewportStyle() {
	if m.debug {
		m.vp.Style = lg.NewStyle().Border(lg.RoundedBorder(), true)
	} else {
		m.vp.Style = lg.NewStyle().
			Padding(2).
			Border(lg.HiddenBorder(), true)
	}
}

func sanitizeStderr(s string) string {
	const maxStderrLen = 500
	if len(s) > maxStderrLen {
		s = s[:maxStderrLen] + "..."
	}
	return strings.TrimSpace(s)
}

// View renders the application UI.
func (m *Model) View() tea.View {
	var out string
	vertView := lg.JoinVertical(0, m.li.View(), m.log.View())
	if m.showConfig {
		out += lg.JoinHorizontal(0.2, vertView, m.vp.View())
	} else {
		out += vertView
	}
	v := tea.NewView(out)
	v.AltScreen = true
	v.WindowTitle = "SSM | Secure Shell Manager"
	v.ReportFocus = true
	v.BackgroundColor = parseHexColor(m.theme.backgroundColor)
	return v
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
