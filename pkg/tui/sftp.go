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
	"sync"
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

type sftpMode int

const (
	modeBrowse sftpMode = iota
	modeConfirmOverwrite
	modeConfirmDelete
	modeMkdir
	modeRename
)

type confirmAction struct {
	mode       sftpMode
	message    string
	localPath  string
	remotePath string
}

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
	name      string
	path      string
	isDir     bool
	isSymlink bool
	size      int64
	modTime   time.Time
	kind      string
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
	previous     *Model
	host         sshconf.Host
	firstBoot    bool
	width        int
	height       int
	activePane   paneSide
	local        filePane
	remote       filePane
	sshCmd       *exec.Cmd
	sshErr       *bytes.Buffer
	sftpClient   *sftp.Client
	status       string
	history      []string
	confirm      *confirmAction
	connecting   *exec.Cmd
	connectMutex sync.Mutex
}

// SftpModel wraps the base model in an SFTP browser sub-model.
func SftpModel(base tea.Model) tea.Model {
	previous, ok := base.(*Model)
	if !ok {
		return base
	}

	i := previous.li.GlobalIndex()
	hosts := previous.config.GetHosts()
	if i < 0 || i >= len(hosts) {
		return base
	}
	host := hosts[i]
	startDir, err := os.Getwd()
	if err != nil {
		startDir, _ = os.UserHomeDir()
	}

	m := &sftpModel{
		previous:   previous,
		host:       host,
		firstBoot:  true,
		activePane: localPane,
		local:      newFilePane("Local", startDir, previous.theme),
		remote:     newFilePane(host.Name, "/tmp", previous.theme),
		status:     "Connecting...",
	}
	m.syncPaneSizes(previous.li.Width(), previous.li.Height())
	return m
}

func newFilePane(title, cwd string, t theme) filePane {
	lightDark := lg.LightDark(true)
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = lg.NewStyle().
		Border(lg.NormalBorder(), false, false, false, true).
		BorderForeground(lightDark(lg.Color("#F79F3F"), lg.Color(t.selectedBorderColor))).
		Foreground(lightDark(lg.Color("#F79F3F"), lg.Color(t.selectedTitleColor))).
		Padding(0, 0, 0, 1)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle.
		Foreground(lightDark(lg.Color("#F79F3F"), lg.Color(t.selectedDescriptionColor)))

	li := list.New([]list.Item{}, delegate, 0, 0)
	li.DisableQuitKeybindings()
	li.Title = ""
	li.SetFilteringEnabled(false)
	li.SetShowHelp(false)
	li.SetShowPagination(false)
	li.SetShowStatusBar(true)
	li.SetStatusBarItemName("file", "files")
	li.Styles.NoItems = lg.NewStyle()
	li.Styles.StatusBar = lg.NewStyle().
		Foreground(lightDark(lg.Color("#A49FA5"), lg.Color("#626262")))

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
			connectRemoteCmd(s.host, s.previous.config.GetPath(), func(cmd *exec.Cmd) {
				s.connectMutex.Lock()
				s.connecting = cmd
				s.connectMutex.Unlock()
			}),
		)
	}

	var cmd tea.Cmd
	if s.activePane == localPane {
		s.local.list, cmd = s.local.list.Update(msg)
	} else {
		s.remote.list, cmd = s.remote.list.Update(msg)
	}
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.syncPaneSizes(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		if s.confirm != nil {
			switch msg.Code {
			case tea.KeyEnter, 'y', 'Y':
				cmd = s.executeConfirm()
				s.confirm = nil
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			case tea.KeyEsc, 'n', 'N':
				s.confirm = nil
			}
			return s, tea.Batch(cmds...)
		}
		switch msg.Code {
		case tea.KeyEsc, 'q':
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
		case 'd':
			cmd = s.handleDelete()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case 'm':
			cmd = s.handleMkdir()
			if cmd != nil {
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
		if s.remote.cwd == "" {
			s.remote.cwd = "/"
		}
		s.status = fmt.Sprintf("Connected to %s", s.host.Name)
		cmds = append(cmds, loadRemoteDirCmd(s.sftpClient, s.remote.cwd))
	case sftpDirMsg:
		if msg.err != nil {
			if msg.side == localPane {
				s.local.list.SetItems([]list.Item{})
			}
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
			s.history = append(s.history, fmt.Sprintf("ERR: %s", msg.err.Error()))
		} else {
			s.status = msg.text
			s.history = append(s.history, msg.text)
		}
		if len(s.history) > 50 {
			s.history = s.history[1:]
		}
		if msg.refreshLocal {
			cmds = append(cmds, loadLocalDirCmd(s.local.cwd))
		}
		if msg.refreshRemote && s.sftpClient != nil {
			cmds = append(cmds, loadRemoteDirCmd(s.sftpClient, s.remote.cwd))
		}
	}

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
		remotePath := pathpkg.Join(s.remote.cwd, filepath.Base(item.path))
		if s.fileExistsRemote(remotePath) {
			s.confirm = &confirmAction{
				mode:       modeConfirmOverwrite,
				message:    fmt.Sprintf("Overwrite %s?", pathpkg.Base(remotePath)),
				localPath:  item.path,
				remotePath: remotePath,
			}
			return nil
		}
		return uploadFileCmd(s.sftpClient, item.path, remotePath)
	case remotePane:
		item, ok := s.remote.list.SelectedItem().(fileItem)
		if !ok {
			return nil
		}
		if item.isDir {
			return loadRemoteDirCmd(s.sftpClient, item.path)
		}
		localPath := filepath.Join(s.local.cwd, pathpkg.Base(item.path))
		if s.fileExistsLocal(localPath) {
			s.confirm = &confirmAction{
				mode:       modeConfirmOverwrite,
				message:    fmt.Sprintf("Overwrite %s?", filepath.Base(localPath)),
				localPath:  localPath,
				remotePath: item.path,
			}
			return nil
		}
		return downloadFileCmd(s.sftpClient, item.path, localPath)
	}
	return nil
}

func (s *sftpModel) handleDelete() tea.Cmd {
	switch s.activePane {
	case localPane:
		item, ok := s.local.list.SelectedItem().(fileItem)
		if !ok || item.name == ".." {
			return nil
		}
		s.confirm = &confirmAction{
			mode:      modeConfirmDelete,
			message:   fmt.Sprintf("Delete %s?", item.name),
			localPath: item.path,
		}
	case remotePane:
		item, ok := s.remote.list.SelectedItem().(fileItem)
		if !ok || item.name == ".." || s.sftpClient == nil {
			return nil
		}
		s.confirm = &confirmAction{
			mode:       modeConfirmDelete,
			message:    fmt.Sprintf("Delete %s?", item.name),
			remotePath: item.path,
		}
	}
	return nil
}

func (s *sftpModel) handleMkdir() tea.Cmd {
	if s.activePane != remotePane || s.sftpClient == nil {
		return transferMsgCmd("", fmt.Errorf("mkdir only available on remote pane"), false, false)
	}
	s.confirm = &confirmAction{
		mode:    modeMkdir,
		message: "mkdir: enter directory name (not yet implemented)",
	}
	return nil
}

func (s *sftpModel) executeConfirm() tea.Cmd {
	if s.confirm == nil {
		return nil
	}
	switch s.confirm.mode {
	case modeConfirmOverwrite:
		if s.confirm.localPath != "" && s.confirm.remotePath != "" {
			if s.activePane == localPane {
				return uploadFileCmd(s.sftpClient, s.confirm.localPath, s.confirm.remotePath)
			}
			return downloadFileCmd(s.sftpClient, s.confirm.remotePath, s.confirm.localPath)
		}
	case modeConfirmDelete:
		if s.confirm.localPath != "" {
			if err := os.RemoveAll(s.confirm.localPath); err != nil {
				return transferMsgCmd("", err, true, false)
			}
			return transferMsgCmd(fmt.Sprintf("deleted %s", filepath.Base(s.confirm.localPath)), nil, true, false)
		}
		if s.confirm.remotePath != "" && s.sftpClient != nil {
			return deleteRemoteFileCmd(s.sftpClient, s.confirm.remotePath)
		}
	}
	return nil
}

func (s *sftpModel) fileExistsLocal(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *sftpModel) fileExistsRemote(path string) bool {
	if s.sftpClient == nil {
		return false
	}
	_, err := s.sftpClient.Stat(path)
	return err == nil
}

func (s *sftpModel) toggleFocus() {
	if s.activePane == localPane {
		s.activePane = remotePane
		return
	}
	s.activePane = localPane
}

func (s *sftpModel) close() {
	s.connectMutex.Lock()
	if s.connecting != nil && s.connecting.Process != nil {
		_ = s.connecting.Process.Kill()
		_, _ = s.connecting.Process.Wait()
		s.connecting = nil
	}
	s.connectMutex.Unlock()

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
	paneHeight := max(10, height-5)

	s.local.list.SetSize(paneWidth, paneHeight)
	s.remote.list.SetSize(paneWidth, paneHeight)
}

func (s *sftpModel) View() tea.View {
	th := s.previous.theme

	remoteStatus := ""
	if s.sftpClient == nil {
		remoteStatus = s.status
	}
	panes := lg.JoinHorizontal(lg.Top,
		s.renderPane(s.local, s.activePane == localPane, ""),
		lg.NewStyle().Foreground(lg.Color(th.mainTitleColor)).Render("│"),
		s.renderPane(s.remote, s.activePane == remotePane, remoteStatus),
	)

	helpBar := s.renderHelp()

	var footer string
	if len(s.history) > 0 {
		show := min(len(s.history), 3)
		recent := s.history[len(s.history)-show:]
		lines := make([]string, 0, len(recent)+1)
		lines = append(lines, lg.NewStyle().Foreground(lg.Color(th.mainTitleColor)).Render("─ transfers ─"))
		for _, h := range recent {
			lines = append(lines, lg.NewStyle().Foreground(lg.Color(th.selectedDescriptionColor)).Render(h))
		}
		footer = "\n" + strings.Join(lines, "\n")
	}

	content := lg.JoinVertical(lg.Left, panes+footer, "", helpBar)

	if s.confirm != nil {
		dialog := s.renderConfirm()
		content = lg.Place(s.width, s.height, lg.Center, lg.Center, dialog)
	}

	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (s *sftpModel) renderHelp() string {
	th := s.previous.theme
	lightDark := lg.LightDark(true)
	keyStyle := lg.NewStyle().Foreground(lightDark(lg.Color("#F79F3F"), lg.Color(th.selectedTitleColor)))
	descStyle := lg.NewStyle().Foreground(lightDark(lg.Color("#A49FA5"), lg.Color(th.selectedDescriptionColor)))

	return keyStyle.Render("enter") + " " + descStyle.Render("transfer") +
		"  •  " + keyStyle.Render("tab") + " " + descStyle.Render("switch pane") +
		"  •  " + keyStyle.Render("←/→") + " " + descStyle.Render("focus") +
		"  •  " + keyStyle.Render("d") + " " + descStyle.Render("delete") +
		"  •  " + keyStyle.Render("q/esc") + " " + descStyle.Render("back")
}

func (s *sftpModel) renderConfirm() string {
	th := s.previous.theme
	box := lg.NewStyle().
		Border(lg.RoundedBorder(), true).
		Padding(1, 2).
		BorderForeground(lg.Color(th.selectedBorderColor))

	message := lg.NewStyle().Bold(true).Render(s.confirm.message)
	prompt := lg.NewStyle().Foreground(lg.Color(th.selectedDescriptionColor)).Render("[y/n]")

	return box.Render(message + "\n\n" + prompt)
}

func (s *sftpModel) renderPane(p filePane, focused bool, status string) string {
	th := s.previous.theme
	headerStyle := lg.NewStyle().
		Bold(true).
		Background(lg.Color(th.mainTitleColor)).
		Foreground(lg.Color("230")).
		Padding(0, 1).
		Width(p.list.Width())
	if focused {
		headerStyle = headerStyle.Background(lg.Color(th.selectedTitleColor))
	}

	headerText := fmt.Sprintf("%s / %s", p.title, p.cwd)
	if status != "" {
		headerText = fmt.Sprintf("%s / %s", p.title, status)
	}
	header := headerStyle.Render(headerText)

	body := lg.NewStyle().
		Padding(0, 1, 0, 0).
		Width(p.list.Width()).
		Render(p.list.View())

	return lg.NewStyle().
		Width(p.list.Width() + 1).
		Render(header + "\n" + body)
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

func connectRemoteCmd(host sshconf.Host, configPath string, onStarted func(*exec.Cmd)) tea.Cmd {
	return func() tea.Msg {
		cmd, stderr, client, root, err := connectSFTP(host, configPath, onStarted)
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
		srcFile, err := os.Open(localPath) //nolint:gosec
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = srcFile.Close() }()

		srcInfo, err := srcFile.Stat()
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		total := srcInfo.Size()

		dst, err := client.Create(remotePath)
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = dst.Close() }()

		name := filepath.Base(localPath)
		pw := &progressWriter{
			writer: dst,
			name:   name,
			total:  total,
		}

		if _, err := io.Copy(pw, srcFile); err != nil {
			return sftpTransferMsg{err: err}
		}

		return sftpTransferMsg{
			text:          fmt.Sprintf("uploaded %s (%s)", name, humanSize(total)),
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

		srcInfo, err := src.Stat()
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		total := srcInfo.Size()

		dst, err := os.Create(localPath) //nolint:gosec
		if err != nil {
			return sftpTransferMsg{err: err}
		}
		defer func() { _ = dst.Close() }()

		name := pathpkg.Base(remotePath)
		pw := &progressWriter{
			writer: dst,
			name:   name,
			total:  total,
		}

		if _, err := io.Copy(pw, src); err != nil {
			return sftpTransferMsg{err: err}
		}

		return sftpTransferMsg{
			text:         fmt.Sprintf("downloaded %s (%s)", name, humanSize(total)),
			refreshLocal: true,
		}
	}
}

type progressWriter struct {
	writer io.Writer
	name   string
	sent   int64
	total  int64
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.sent += int64(n)
	return n, err
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

func deleteRemoteFileCmd(client *sftp.Client, path string) tea.Cmd {
	return func() tea.Msg {
		if err := client.Remove(path); err != nil {
			return sftpTransferMsg{err: err, refreshRemote: true}
		}
		return sftpTransferMsg{
			text:          fmt.Sprintf("deleted %s", pathpkg.Base(path)),
			refreshRemote: true,
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
		isSymlink := entry.Type()&os.ModeSymlink != 0
		items = append(items, fileItem{
			name:      entry.Name(),
			path:      filepath.Join(path, entry.Name()),
			isDir:     entry.IsDir(),
			isSymlink: isSymlink,
			size:      info.Size(),
			modTime:   info.ModTime(),
			kind:      fileKind(entry.Name(), entry.IsDir(), isSymlink),
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
		isSymlink := entry.Mode()&os.ModeSymlink != 0
		items = append(items, fileItem{
			name:      entry.Name(),
			path:      pathpkg.Join(path, entry.Name()),
			isDir:     entry.IsDir(),
			isSymlink: isSymlink,
			size:      entry.Size(),
			modTime:   entry.ModTime(),
			kind:      fileKind(entry.Name(), entry.IsDir(), isSymlink),
		})
	}
	return items, nil
}

func connectSFTP(host sshconf.Host, configPath string, onStarted func(*exec.Cmd)) (*exec.Cmd, *bytes.Buffer, *sftp.Client, string, error) {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("ssh not found in PATH: %w", err)
	}

	//nolint:gosec
	// StrictHostKeyChecking=no is intentional — users connect to their own
	// servers by name. Do not change this without explicit project approval.
	cmd := exec.Command(sshPath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "BatchMode=yes",
		"-o", "RequestTTY=no",
		"-s",
		"-F", configPath,
		"--", host.Name, "sftp",
	)
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

	if onStarted != nil {
		onStarted(cmd)
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
		root = "/tmp"
	}

	return cmd, stderr, client, root, nil
}

func fileKind(name string, isDir, isSymlink bool) string {
	if isSymlink {
		return "link"
	}
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
