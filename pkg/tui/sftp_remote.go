package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/pkg/sftp"
)

type remoteFile struct {
	Name  string
	IsDir bool
}

func (f remoteFile) Title() string {
	return f.Name
}

func (f remoteFile) Description() string {
	if f.IsDir {
		return "directory"
	}
	return "file"
}

func (f remoteFile) FilterValue() string {
	return f.Name
}

type remoteFileSystem struct {
	list      list.Model
	client    *sftp.Client
	cwd       string
	host      sshconf.Host
	selected  string
	connected bool
}

func NewRemoteFileSystem(host *sshconf.Host) tea.Model {

	return &remoteFileSystem{
		cwd:  ".",
		host: *host,
	}
}

func (l *remoteFileSystem) Init() tea.Cmd {
	return nil
}

func (l *remoteFileSystem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return l, nil
		}

	}

	return l, nil
}

func (l *remoteFileSystem) View() string {

	return ""
}

func loadDir(client *sftp.Client, path string) []list.Item {
	entries, err := client.ReadDir(path)
	if err != nil {
		return []list.Item{remoteFile{Name: fmt.Sprintf("error: %v", err)}}
	}

	items := []list.Item{}
	if path != "/" {
		items = append(items, remoteFile{Name: "..", IsDir: true}) // parent dir
	}
	for _, e := range entries {
		items = append(items, remoteFile{Name: e.Name(), IsDir: e.IsDir()})
	}
	return items
}
