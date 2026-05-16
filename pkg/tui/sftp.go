package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/pkg/sftp"
)

type paneSide int

const (
	localPane paneSide = iota
	remotePane
)

type sftpConnectMsg struct {
	cmd    *exec.Cmd
	stderr *bytes.Buffer
	client *sftp.Client
	root   string
	err    error
}

type sftpDirMsg struct {
	side  paneSide
	path  string
	items []list.Item
	err   error
}

type sftpTransferMsg struct {
	text          string
	err           error
	refreshLocal  bool
	refreshRemote bool
}

type fileItem struct {
	name    string
	path    string
	isDir   bool
	size    int64
	modTime time.Time
	kind    string
}

func (f fileItem) Title() string {
	return f.name
}

func (f fileItem) Description() string {
	if f.name == ".." {
		return "parent directory"
	}

	size := "--"
	if !f.isDir {
		size = humanSize(f.size)
	}
	stamp := "--"
	if !f.modTime.IsZero() {
		stamp = f.modTime.Format("2006-01-02 15:04")
	}
	return fmt.Sprintf("%-16s  %-8s  %s", stamp, size, f.kind)
}

func (f fileItem) FilterValue() string {
	return f.name
}

type filePane struct {
	title string
	cwd   string
	list  list.Model
}

type sftpModel struct {
	previous   *Model
	host       sshconf.Host
	firstBoot  bool
	width      int
	height     int
	activePane paneSide
	local      filePane
	remote     filePane
	sshCmd     *exec.Cmd
	sshErr     *bytes.Buffer
	sftpClient *sftp.Client
	status     string
}

// SftpModel wraps the base model in an SFTP browser sub-model.
func SftpModel(base tea.Model) tea.Model {
	previous, ok := base.(*Model)
	if !ok {
		panic("failed to cast tea.Model to Model")
	}

	i := previous.li.GlobalIndex()
	host := previous.config.Hosts[i]
	startDir, err := os.Getwd()
	if err != nil {
		startDir, _ = os.UserHomeDir()
	}

	m := &sftpModel{
		previous:   previous,
		host:       host,
		firstBoot:  true,
		activePane: localPane,
		local:      newFilePane("Local", startDir),
		remote:     newFilePane(host.Name, "."),
		status:     "Connecting...",
	}
	m.syncPaneSizes(previous.li.Width(), previous.li.Height())
	return m
}

func newFilePane(title, cwd string) filePane {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.ShowDescription = true

	li := list.New([]list.Item{}, delegate, 0, 0)
	li.DisableQuitKeybindings()
	li.SetFilteringEnabled(false)
	li.SetShowHelp(false)
	li.SetShowPagination(false)
	li.SetShowStatusBar(false)
	li.SetShowFilter(false)

	return filePane{
		title: title,
		cwd:   cwd,
		list:  li,
	}
}

func (s *sftpModel) Init() tea.Cmd {
	return nil
}

func (s *sftpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if s.firstBoot {
		s.firstBoot = false
		cmds = append(cmds,
			loadLocalDirCmd(s.local.cwd),
			connectRemoteCmd(s.host, s.previous.config.GetPath()),
		)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.syncPaneSizes(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEsc:
			s.close()
			return s.previous, nil
		case tea.KeyTab:
			s.toggleFocus()
		case tea.KeyLeft:
			s.activePane = localPane
		case tea.KeyRight:
			s.activePane = remotePane
		case tea.KeyEnter:
			if cmd := s.handleEnter(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case sftpConnectMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
			return s, tea.Batch(cmds...)
		}
		s.sshCmd = msg.cmd
		s.sshErr = msg.stderr
		s.sftpClient = msg.client
		s.remote.cwd = msg.root
		s.status = fmt.Sprintf("Connected to %s", s.host.Name)
		cmds = append(cmds, loadRemoteDirCmd(s.sftpClient, s.remote.cwd))
	case sftpDirMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
			break
		}
		if msg.side == localPane {
			s.local.cwd = msg.path
			s.local.list.SetItems(msg.items)
		} else {
			s.remote.cwd = msg.path
			s.remote.list.SetItems(msg.items)
		}
	case sftpTransferMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
		} else {
			s.status = msg.text
		}
		if msg.refreshLocal {
			cmds = append(cmds, loadLocalDirCmd(s.local.cwd))
		}
		if msg.refreshRemote && s.sftpClient != nil {
			cmds = append(cmds, loadRemoteDirCmd(s.sftpClient, s.remote.cwd))
		}
	}

	var cmd tea.Cmd
	if s.activePane == localPane {
		s.local.list, cmd = s.local.list.Update(msg)
	} else {
		s.remote.list, cmd = s.remote.list.Update(msg)
	}
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s *sftpModel) handleEnter() tea.Cmd {
	switch s.activePane {
	case localPane:
		item, ok := s.local.list.SelectedItem().(fileItem)
		if !ok {
			return nil
		}
		if item.isDir {
			return loadLocalDirCmd(item.path)
		}
		if s.sftpClient == nil {
			return transferMsgCmd("", fmt.Errorf("not connected to remote host"), false, false)
		}
		return uploadFileCmd(s.sftpClient, item.path, pathpkg.Join(s.remote.cwd, filepath.Base(item.path)))
	case remotePane:
		item, ok := s.remote.list.SelectedItem().(fileItem)
		if !ok {
			return nil
		}
		if item.isDir {
			return loadRemoteDirCmd(s.sftpClient, item.path)
		}
		return downloadFileCmd(s.sftpClient, item.path, filepath.Join(s.local.cwd, pathpkg.Base(item.path)))
	}
	return nil
}

func (s *sftpModel) toggleFocus() {
	if s.activePane == localPane {
		s.activePane = remotePane
		return
	}
	s.activePane = localPane
}

func (s *sftpModel) close() {
	if s.sftpClient != nil {
		_ = s.sftpClient.Close()
		s.sftpClient = nil
	}
	if s.sshCmd != nil && s.sshCmd.Process != nil {
		_ = s.sshCmd.Process.Kill()
		_, _ = s.sshCmd.Process.Wait()
		s.sshCmd = nil
	}
}

func (s *sftpModel) syncPaneSizes(width, height int) {
	if width <= 0 {
		width = s.previous.li.Width()
	}
	if height <= 0 {
		height = s.previous.li.Height()
	}
	s.width = width
	s.height = height

	paneWidth := max(20, (width/2)-2)
	paneHeight := max(10, height-6)

	s.local.list.SetSize(paneWidth, paneHeight)
	s.remote.list.SetSize(paneWidth, paneHeight)
}

func (s *sftpModel) View() tea.View {
	v := lg.JoinVertical(
		lg.Left,
		lg.NewStyle().
			Foreground(lg.Color("8")).
			Render(fmt.Sprintf("sftp  %s  |  tab switch  |  enter open/transfer  |  esc back", s.status)),
		"",
		lg.JoinHorizontal(lg.Top,
			s.renderPane(s.local, s.activePane == localPane),
			lg.NewStyle().Foreground(lg.Color("8")).Render("│"),
			s.renderPane(s.remote, s.activePane == remotePane),
		),
	)
	view := tea.NewView(v)
	view.AltScreen = true
	return view
}

func (s *sftpModel) renderPane(p filePane, focused bool) string {
	headerStyle := lg.NewStyle().
		Bold(focused).
		Foreground(lg.Color("8"))
	if focused {
		headerStyle = headerStyle.Foreground(lg.Color(s.previous.theme.selectedTitleColor))
	}

	header := headerStyle.Render(fmt.Sprintf("%s  %s", p.title, p.cwd))
	underline := lg.NewStyle().
		Foreground(lg.Color("8")).
		Render(strings.Repeat("─", max(1, p.list.Width())))
	if focused {
		underline = lg.NewStyle().
			Foreground(lg.Color(s.previous.theme.selectedBorderColor)).
			Render(strings.Repeat("─", max(1, p.list.Width())))
	}

	body := lg.NewStyle().
		Padding(0, 1, 0, 0).
		Width(p.list.Width()).
		Render(p.list.View())

	return lg.NewStyle().
		Width(p.list.Width() + 1).
		Render(header + "\n" + underline + "\n" + body)
}

func loadLocalDirCmd(path string) tea.Cmd {
	return func() tea.Msg {
		items, err := loadLocalDir(path)
		return sftpDirMsg{
			side:  localPane,
			path:  path,
			items: items,
			err:   err,
		}
	}
}

func loadRemoteDirCmd(client *sftp.Client, path string) tea.Cmd {
	return func() tea.Msg {
		items, err := loadRemoteDir(client, path)
		return sftpDirMsg{
			side:  remotePane,
			path:  path,
			items: items,
			err:   err,
		}
	}
}

func connectRemoteCmd(host sshconf.Host, configPath string) tea.Cmd {
	return func() tea.Msg {
		cmd, stderr, client, root, err := connectSFTP(host, configPath)
		return sftpConnectMsg{
			cmd:    cmd,
			stderr: stderr,
			client: client,
			root:   root,
			err:    err,
		}
	}
}

func uploadFileCmd(client *sftp.Client, localPath, remotePath string) tea.Cmd {
	return func() tea.Msg {
		src, err := os.Open(localPath) //nolint:gosec
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = src.Close() }()

		dst, err := client.Create(remotePath)
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = dst.Close() }()

		if _, err := io.Copy(dst, src); err != nil {
			return sftpTransferMsg{err: err}
		}

		return sftpTransferMsg{
			text:          fmt.Sprintf("uploaded %s", filepath.Base(localPath)),
			refreshRemote: true,
		}
	}
}

func downloadFileCmd(client *sftp.Client, remotePath, localPath string) tea.Cmd {
	return func() tea.Msg {
		src, err := client.Open(remotePath)
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = src.Close() }()

		dst, err := os.Create(localPath) //nolint:gosec
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = dst.Close() }()

		if _, err := io.Copy(dst, src); err != nil {
			return sftpTransferMsg{err: err}
		}

		return sftpTransferMsg{
			text:         fmt.Sprintf("downloaded %s", pathpkg.Base(remotePath)),
			refreshLocal: true,
		}
	}
}

func transferMsgCmd(text string, err error, refreshLocal, refreshRemote bool) tea.Cmd {
	return func() tea.Msg {
		return sftpTransferMsg{
			text:          text,
			err:           err,
			refreshLocal:  refreshLocal,
			refreshRemote: refreshRemote,
		}
	}
}

func loadLocalDir(path string) ([]list.Item, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, 0, len(entries)+1)
	parent := filepath.Dir(path)
	if parent != path {
		items = append(items, fileItem{
			name:  "..",
			path:  parent,
			isDir: true,
			kind:  "dir",
		})
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, fileItem{
			name:    entry.Name(),
			path:    filepath.Join(path, entry.Name()),
			isDir:   entry.IsDir(),
			size:    info.Size(),
			modTime: info.ModTime(),
			kind:    fileKind(entry.Name(), entry.IsDir()),
		})
	}
	return items, nil
}

func loadRemoteDir(client *sftp.Client, path string) ([]list.Item, error) {
	entries, err := client.ReadDir(path)
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, 0, len(entries)+1)
	parent := pathpkg.Dir(path)
	if parent != path {
		items = append(items, fileItem{
			name:  "..",
			path:  parent,
			isDir: true,
			kind:  "dir",
		})
	}

	for _, entry := range entries {
		items = append(items, fileItem{
			name:    entry.Name(),
			path:    pathpkg.Join(path, entry.Name()),
			isDir:   entry.IsDir(),
			size:    entry.Size(),
			modTime: entry.ModTime(),
			kind:    fileKind(entry.Name(), entry.IsDir()),
		})
	}
	return items, nil
}

func connectSFTP(host sshconf.Host, configPath string) (*exec.Cmd, *bytes.Buffer, *sftp.Client, string, error) {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("ssh not found in PATH: %w", err)
	}

	cmd := exec.Command(sshPath, "-F", configPath, host.Name, "-s", "sftp") //nolint:gosec
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, "", err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, "", err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	client, err := sftp.NewClientPipe(stdout, stdin)
	if err != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, nil, nil, "", fmt.Errorf("%w: %s", err, msg)
		}
		return nil, nil, nil, "", err
	}

	root, err := client.Getwd()
	if err != nil || root == "" {
		root = "."
	}

	return cmd, stderr, client, root, nil
}

func fileKind(name string, isDir bool) string {
	if isDir {
		return "dir"
	}
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	if ext == "" {
		return "file"
	}
	return ext
}

func humanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(size) / float64(div)
	return strconv.FormatFloat(value, 'f', 1, 64) + string("KMGTPE"[exp]) + "B"
}
