package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/v2/filepicker"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	// "github.com/charmbracelet/lipgloss/v2"
)

type localFileSystem struct {
	filePicker   filepicker.Model
	selectedFile string
	viewport     viewport.Model
}

func NewLocalFileSystem() tea.Model {

	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()

	vp := viewport.New()

	return &localFileSystem{
		filePicker: fp,
		viewport:   vp,
	}
}

func (l *localFileSystem) Init() tea.Cmd {
	return l.filePicker.Init()
}

func (l *localFileSystem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return l, nil
		}

	case tea.WindowSizeMsg:
		l.filePicker.SetHeight(msg.Height)
		l.viewport.SetHeight(msg.Height)
		l.viewport.SetWidth(msg.Width)
		// l.viewport.Style = lipgloss.NewStyle().Background(lipgloss.Color("#28f3433"))

	}

	var cmd tea.Cmd
	l.filePicker, cmd = l.filePicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := l.filePicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		l.selectedFile = path
	}

	return l, cmd
}

func (l *localFileSystem) View() string {

	l.viewport.SetContent(l.filePicker.View())
	// return l.filePicker.View()
	return l.viewport.View()
}
