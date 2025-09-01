package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type sftpModel struct {
	previous      tea.Model
	remoteFSModel tea.Model
	localFSModel  tea.Model
	fistBoot      bool
}

func SftpModel(base tea.Model) tea.Model {

	previous, ok := base.(*Model)

	i := previous.li.GlobalIndex()
	host := previous.config.Hosts[i]

	if !ok {
		panic("failed to cast tea.Model to Model")
	}

	return &sftpModel{
		previous:      base,
		fistBoot:      true,
		localFSModel:  NewLocalFileSystem(),
		remoteFSModel: NewRemoteFileSystem(&host),
	}
}

func (s *sftpModel) Init() tea.Cmd {
	return nil
}

func (s *sftpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmds []tea.Cmd
	var localFSCmd tea.Cmd
	var remoteFSCmd tea.Cmd

	// this is a workaround because Init is not called
	if s.fistBoot {
		s.fistBoot = false

		previousModel, ok := s.previous.(*Model)
		if !ok {
			panic("failed to cast tea.Model to Model")
		}

		var windowSize tea.WindowSizeMsg

		windowSize.Width = previousModel.vp.Width()
		windowSize.Height = previousModel.vp.Height()

		var uLocalFSCmd, uRemoteFSCmd tea.Cmd
		s.localFSModel, _ = s.localFSModel.Update(windowSize)
		s.remoteFSModel, _ = s.remoteFSModel.Update(windowSize)

		cmds = append(cmds, uLocalFSCmd, uRemoteFSCmd, s.localFSModel.Init(), s.remoteFSModel.Init())

		// return s, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s.previous, nil
		}

	}

	s.localFSModel, localFSCmd = s.localFSModel.Update(msg)
	s.remoteFSModel, remoteFSCmd = s.remoteFSModel.Update(msg)

	cmds = append(cmds, localFSCmd, remoteFSCmd)

	return s, tea.Batch(cmds...)
}

func (s *sftpModel) View() string {

	local, ok := s.localFSModel.(*localFileSystem)
	remote, ok := s.remoteFSModel.(*remoteFileSystem)

	if !ok {
		panic("failed to cast tea.Model to localFileSystem Model")
	}

	var view strings.Builder
	view.WriteString("send receive files")

	fs := lipgloss.JoinHorizontal(lipgloss.Top, local.View(), remote.View())
	view.WriteString("\n\n" + fs + "\n")

	return view.String()
}
