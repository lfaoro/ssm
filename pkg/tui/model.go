package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	lp "github.com/charmbracelet/lipgloss/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
)

type Model struct {
	config *sshconf.Config
	li     list.Model
	sshCmd string

	debug bool
	log   Log

	errbuf bytes.Buffer
	isDark bool
}

type ReloadConfigMsg struct{}

func NewModel(config *sshconf.Config, debug bool) *Model {
	m := &Model{}
	m.debug = debug
	m.config = config
	m.li = listFrom(config)
	m.log = NewLog(WithDebug(debug))
	return m
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.SetWindowTitle("SSM | Secure Shell Manager"),
		// tea.SetBackgroundColor(color.Black),
		tea.RequestKeyboardEnhancements(),
		tea.EnterAltScreen,
		tea.EnableBracketedPaste,
		tea.EnableReportFocus,
		// tea.EnableMouseAllMotion,
		// tea.EnableMouseCellMotion,
	}
	if m.debug {
		cmds = append(cmds, AddLog("debug: isdarkbg %v", m.isDark))
	}
	m.sshCmd = "ssh"
	m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.sshCmd))

	// reload config on edit
	cmds = append(cmds, m.watchCmd())

	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case ReloadConfigMsg:
		m.li = listFrom(m.config)
		return m, nil

	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()

	case tea.WindowSizeMsg:
		m.li.SetSize(msg.Width, msg.Height-3)
		if m.debug {
			m.li.SetSize(msg.Width, msg.Height-9)
		}

	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyTab:
			if m.sshCmd == "ssh" {
				m.sshCmd = "mosh"
				m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.sshCmd))
			} else {
				m.sshCmd = "ssh"
				m.li.NewStatusMessage(fmt.Sprintf("[%s]", m.sshCmd))
			}

		case tea.KeyEnter:
			// connect
			host, ok := m.li.SelectedItem().(item)
			if !ok {
				return m, AddError(fmt.Errorf("unable to find selected item: open bug report"))
			}
			sshcmd, err := exec.LookPath(m.sshCmd)
			if err != nil {
				return m, AddError(fmt.Errorf("can't find `%s` cmd in your path: %v", m.sshCmd, err))
			}
			var cmd *exec.Cmd
			cmd = exec.Command(sshcmd, host.title)
			if host.title == "segfault.net" {
				_sshcmd, err := exec.LookPath("sshpass")
				if err != nil {
					_sshcmd, err = exec.LookPath("ssh")
					if err != nil {
						return m, AddError(fmt.Errorf("can't find `%s` cmd in your path: %v", m.sshCmd, err))
					}
					cmd = exec.Command(_sshcmd, "root@segfault.net")
				} else {
					cmd = exec.Command(_sshcmd, "-p", "segfault", "ssh", "root@segfault.net")
				}
			}
			cmd.Stderr = &m.errbuf
			execmd := tea.ExecProcess(cmd, func(err error) tea.Msg {
				return tea.Batch(
					AddError(
						fmt.Errorf("connection closed: %v, err: %v", host.title, err),
					),
					AddError(fmt.Errorf("%s", m.errbuf.String())),
				)
			})
			return m, tea.Batch(
				execmd,
				AddError(fmt.Errorf("%s", m.errbuf.String())),
			)
		case 'c':
		// run command
		case tea.KeyEscape:
		case 'q':
			if m.li.FilterState() != 1 {
				return m, tea.Quit
			}
		}
		switch msg.Mod {
		// we're only interested in ctrl+<key>
		case tea.ModCtrl:
			switch msg.Code {
			case 'e':
				confFile := m.config.GetPath()
				cmd := exec.Command(os.Getenv("EDITOR"), confFile)
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
			case 'c':
				return m, AddError(fmt.Errorf("run command on host: not yet implemented"))
			case 's':
				return m, AddError(fmt.Errorf("sftp: not yet implemented"))
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

	m.log, cmd = m.log.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var out string
	style := lp.NewStyle().
		Margin(1, 0, 0, 1).
		Render
	out += style(m.li.View())
	out += style(m.log.View())
	return out
}

func (m *Model) watchCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.config.Watch()
		if err != nil {
			return tea.Batch(
				AddError(err),
				AddLog("%v", err),
			)
		}
		return nil
	}
}
