// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func TestNewFilePane(t *testing.T) {
	t.Run("creates pane with defaults", func(t *testing.T) {
		p := newFilePane("Test", "/tmp", matrixTheme())

		if p.title != "Test" {
			t.Errorf("title = %q, want %q", p.title, "Test")
		}
		if p.cwd != "/tmp" {
			t.Errorf("cwd = %q, want %q", p.cwd, "/tmp")
		}
	})
}

func TestFileItem_Title(t *testing.T) {
	tests := []struct {
		name string
		item fileItem
		want string
	}{
		{
			name: "regular file",
			item: fileItem{name: "test.txt"},
			want: "test.txt",
		},
		{
			name: "directory",
			item: fileItem{name: "src", isDir: true},
			want: "src",
		},
		{
			name: "symlink",
			item: fileItem{name: "link", isSymlink: true},
			want: "link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.Title()
			if got != tt.want {
				t.Errorf("Title() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFileItem_Description(t *testing.T) {
	t.Run("parent directory", func(t *testing.T) {
		item := fileItem{name: ".."}
		got := item.Description()
		if got != "parent directory" {
			t.Errorf("Description() = %q, want %q", got, "parent directory")
		}
	})

	t.Run("directory shows dir kind", func(t *testing.T) {
		item := fileItem{name: "src", isDir: true, kind: "dir"}
		got := item.Description()
		if got == "" {
			t.Error("expected non-empty description for directory")
		}
	})

	t.Run("file shows size and kind", func(t *testing.T) {
		item := fileItem{
			name:    "main.go",
			size:    1024,
			modTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			kind:    "go",
		}
		got := item.Description()
		if got == "" {
			t.Error("expected non-empty description for file")
		}
	})
}

func TestFileItem_FilterValue(t *testing.T) {
	item := fileItem{name: "test.txt"}
	got := item.FilterValue()
	if got != "test.txt" {
		t.Errorf("FilterValue() = %q, want %q", got, "test.txt")
	}
}

func TestFileKind(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		isDir     bool
		isSymlink bool
		want      string
	}{
		{"directory", "src", true, false, "dir"},
		{"symlink", "link", false, true, "link"},
		{"symlink dir", "linkdir", true, true, "link"},
		{"file with extension", "main.go", false, false, "go"},
		{"file without extension", "Makefile", false, false, "file"},
		{"hidden file", ".gitignore", false, false, "gitignore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileKind(tt.fileName, tt.isDir, tt.isSymlink)
			if got != tt.want {
				t.Errorf("fileKind(%q, %v, %v) = %q, want %q", tt.fileName, tt.isDir, tt.isSymlink, got, tt.want)
			}
		})
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"zero bytes", 0, "0B"},
		{"bytes", 512, "512B"},
		{"kilobytes", 1024, "1.0KB"},
		{"megabytes", 1048576, "1.0MB"},
		{"gigabytes", 1073741824, "1.0GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanSize(tt.size)
			if got != tt.want {
				t.Errorf("humanSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestLoadLocalDir(t *testing.T) {
	t.Run("valid directory", func(t *testing.T) {
		tmp := t.TempDir()
		_ = os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("test"), 0644) //nolint:gosec
		_ = os.Mkdir(filepath.Join(tmp, "subdir"), 0755)                       //nolint:gosec

		items, err := loadLocalDir(tmp)
		if err != nil {
			t.Fatalf("loadLocalDir() error = %v", err)
		}

		if len(items) < 3 {
			t.Errorf("expected at least 3 items (.., file.txt, subdir), got %d", len(items))
		}

		found := false
		for _, it := range items {
			fi, ok := it.(fileItem)
			if !ok {
				continue
			}
			if fi.name == "file.txt" {
				found = true
				if fi.isDir {
					t.Error("file.txt should not be a directory")
				}
				if fi.kind != "txt" {
					t.Errorf("file.txt kind = %q, want %q", fi.kind, "txt")
				}
			}
			if fi.name == "subdir" {
				if !fi.isDir {
					t.Error("subdir should be a directory")
				}
				if fi.kind != "dir" {
					t.Errorf("subdir kind = %q, want %q", fi.kind, "dir")
				}
			}
		}
		if !found {
			t.Error("file.txt not found in items")
		}
	})

	t.Run("parent directory entry", func(t *testing.T) {
		tmp := t.TempDir()
		items, err := loadLocalDir(tmp)
		if err != nil {
			t.Fatalf("loadLocalDir() error = %v", err)
		}

		found := false
		for _, it := range items {
			fi, ok := it.(fileItem)
			if !ok {
				continue
			}
			if fi.name == ".." {
				found = true
				if !fi.isDir {
					t.Error(".. should be a directory")
				}
			}
		}
		if !found {
			t.Error(".. entry not found")
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := loadLocalDir("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Error("expected error for nonexistent directory")
		}
	})
}

func TestSftpModel_Init(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp, ok := result.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result)
	}

	cmd := sftp.Init()
	if cmd != nil {
		t.Error("expected nil command from Init() — submodels use firstBoot in Update()")
	}
}

func TestSftpModel_FirstBoot(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp, ok := result.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result)
	}

	if !sftp.firstBoot {
		t.Error("expected firstBoot to be true on creation")
	}

	result2, _ := sftp.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	sftp2, ok := result2.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result2)
	}

	if sftp2.firstBoot {
		t.Error("expected firstBoot to be false after first Update()")
	}
}

func TestSftpModel_Update_WindowSize(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp, ok := result.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result)
	}

	sftp.width = 100
	sftp.height = 30
	result2, _ := sftp.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	sftp2 := result2.(*sftpModel)

	if sftp2.width != 120 {
		t.Errorf("width = %d, want 120", sftp2.width)
	}
	if sftp2.height != 40 {
		t.Errorf("height = %d, want 40", sftp2.height)
	}
}

func TestSftpModel_Update_Esc_ReturnsToPrevious(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp, ok := result.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result)
	}

	result2, _ := sftp.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if result2 != m {
		t.Error("expected to return to previous model on Esc")
	}
}

func TestSftpModel_Update_Tab_ToggleFocus(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp, ok := result.(*sftpModel)
	if !ok {
		t.Fatalf("expected *sftpModel, got %T", result)
	}

	if sftp.activePane != localPane {
		t.Errorf("initial activePane = %v, want %v", sftp.activePane, localPane)
	}

	result2, _ := sftp.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sftp2 := result2.(*sftpModel)

	if sftp2.activePane != remotePane {
		t.Errorf("after tab activePane = %v, want %v", sftp2.activePane, remotePane)
	}

	result3, _ := sftp2.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sftp3 := result3.(*sftpModel)

	if sftp3.activePane != localPane {
		t.Errorf("after second tab activePane = %v, want %v", sftp3.activePane, localPane)
	}
}

func TestSftpModel_Update_LeftRight_PaneSwitch(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	result2, _ := sftp.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	sftp2 := result2.(*sftpModel)

	if sftp2.activePane != remotePane {
		t.Errorf("after right key activePane = %v, want %v", sftp2.activePane, remotePane)
	}

	result3, _ := sftp2.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	sftp3 := result3.(*sftpModel)

	if sftp3.activePane != localPane {
		t.Errorf("after left key activePane = %v, want %v", sftp3.activePane, localPane)
	}
}

func TestSftpModel_toggleFocus(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.toggleFocus()
	if sftp.activePane != remotePane {
		t.Errorf("after toggle activePane = %v, want %v", sftp.activePane, remotePane)
	}

	sftp.toggleFocus()
	if sftp.activePane != localPane {
		t.Errorf("after second toggle activePane = %v, want %v", sftp.activePane, localPane)
	}
}

func TestSftpModel_close(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.close()

	if sftp.sftpClient != nil {
		t.Error("expected sftpClient to be nil after close")
	}
}

func TestSftpModel_handleEnter_NoSelection(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	cmd := sftp.handleEnter()
	if cmd != nil {
		t.Error("expected nil command when no item selected")
	}
}

func TestSftpModel_handleEnter_LocalDirectory(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	tmp := t.TempDir()
	sftp.local.list.SetItems([]list.Item{
		fileItem{name: "subdir", path: tmp, isDir: true, kind: "dir"},
	})
	sftp.local.list.Select(0)

	cmd := sftp.handleEnter()
	if cmd == nil {
		t.Error("expected command for directory navigation")
	}
}

func TestSftpModel_handleEnter_RemoteDirectory(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.activePane = remotePane
	sftp.remote.list.SetItems([]list.Item{
		fileItem{name: "remoteDir", path: "/home/user/docs", isDir: true, kind: "dir"},
	})
	sftp.remote.list.Select(0)

	cmd := sftp.handleEnter()
	if cmd == nil {
		t.Error("expected command for remote directory navigation")
	}
}

func TestSftpModel_handleEnter_Upload_NoConnection(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("test"), 0644) //nolint:gosec
	sftp.local.list.SetItems([]list.Item{
		fileItem{name: "file.txt", path: filepath.Join(tmp, "file.txt"), isDir: false, kind: "txt"},
	})
	sftp.local.list.Select(0)

	cmd := sftp.handleEnter()
	if cmd == nil {
		t.Fatal("expected command when uploading without connection")
	}

	msg := cmd()
	transferMsg, ok := msg.(sftpTransferMsg)
	if !ok {
		t.Fatalf("expected sftpTransferMsg, got %T", msg)
	}
	if transferMsg.err == nil {
		t.Error("expected error when not connected")
	}
}

func TestSftpModel_sftpTransferMsg_Error(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	testErr := os.ErrNotExist
	result2, _ := sftp.Update(sftpTransferMsg{err: testErr})
	sftp2 := result2.(*sftpModel)

	if sftp2.status != testErr.Error() {
		t.Errorf("status = %q, want %q", sftp2.status, testErr.Error())
	}
	if len(sftp2.history) != 1 {
		t.Errorf("history length = %d, want 1", len(sftp2.history))
	}
}

func TestSftpModel_sftpTransferMsg_Success(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	result2, _ := sftp.Update(sftpTransferMsg{
		text:          "uploaded file.txt (1.0KB)",
		refreshRemote: true,
	})
	sftp2 := result2.(*sftpModel)

	if sftp2.status != "uploaded file.txt (1.0KB)" {
		t.Errorf("status = %q, want %q", sftp2.status, "uploaded file.txt (1.0KB)")
	}
	if len(sftp2.history) != 1 {
		t.Errorf("history length = %d, want 1", len(sftp2.history))
	}
	if sftp2.history[0] != "uploaded file.txt (1.0KB)" {
		t.Errorf("history[0] = %q, want %q", sftp2.history[0], "uploaded file.txt (1.0KB)")
	}
}

func TestSftpModel_sftpDirMsg_Local(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	items := []list.Item{
		fileItem{name: "file.txt", path: "/tmp/file.txt", isDir: false, kind: "txt"},
	}
	result2, _ := sftp.Update(sftpDirMsg{
		side:  localPane,
		path:  "/tmp",
		items: items,
	})
	sftp2 := result2.(*sftpModel)

	if sftp2.local.cwd != "/tmp" {
		t.Errorf("local.cwd = %q, want %q", sftp2.local.cwd, "/tmp")
	}
	if len(sftp2.local.list.Items()) != 1 {
		t.Errorf("local list items = %d, want 1", len(sftp2.local.list.Items()))
	}
}

func TestSftpModel_sftpDirMsg_Remote(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	items := []list.Item{
		fileItem{name: "remote.txt", path: "/home/remote.txt", isDir: false, kind: "txt"},
	}
	result2, _ := sftp.Update(sftpDirMsg{
		side:  remotePane,
		path:  "/home",
		items: items,
	})
	sftp2 := result2.(*sftpModel)

	if sftp2.remote.cwd != "/home" {
		t.Errorf("remote.cwd = %q, want %q", sftp2.remote.cwd, "/home")
	}
}

func TestSftpModel_ConfirmDialog(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.confirm = &confirmAction{
		mode:    modeConfirmOverwrite,
		message: "Overwrite test.txt?",
	}

	v := sftp.View()
	if v.AltScreen == false {
		t.Error("expected AltScreen to be true")
	}

	result2, _ := sftp.Update(tea.KeyPressMsg{Code: 'n'})
	sftp2 := result2.(*sftpModel)

	if sftp2.confirm != nil {
		t.Error("expected confirm to be cleared after 'n'")
	}
}

func TestSftpModel_ConfirmDialog_Accept(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.confirm = &confirmAction{
		mode:    modeConfirmDelete,
		message: "Delete test.txt?",
	}

	result2, _ := sftp.Update(tea.KeyPressMsg{Code: 'y'})
	sftp2 := result2.(*sftpModel)

	if sftp2.confirm != nil {
		t.Error("expected confirm to be cleared after 'y'")
	}
}

func TestSftpModel_handleDelete_LocalPane(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "del.txt"), []byte("test"), 0644) //nolint:gosec
	sftp.local.list.SetItems([]list.Item{
		fileItem{name: "del.txt", path: filepath.Join(tmp, "del.txt"), isDir: false, kind: "txt"},
	})
	sftp.local.list.Select(0)

	sftp.handleDelete()

	if sftp.confirm == nil {
		t.Fatal("expected confirm dialog for delete")
	}
	if sftp.confirm.mode != modeConfirmDelete {
		t.Errorf("confirm.mode = %v, want %v", sftp.confirm.mode, modeConfirmDelete)
	}
}

func TestSftpModel_handleDelete_ParentEntry(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.local.list.SetItems([]list.Item{
		fileItem{name: "..", path: "/tmp", isDir: true, kind: "dir"},
	})
	sftp.local.list.Select(0)

	sftp.handleDelete()

	if sftp.confirm != nil {
		t.Error("expected no confirm dialog for .. entry")
	}
}

func TestSftpModel_handleMkdir_RemotePane(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.activePane = remotePane
	sftp.sftpClient = nil

	cmd := sftp.handleMkdir()
	if cmd == nil {
		t.Error("expected command for mkdir on remote without client")
	}
}

func TestSftpModel_handleMkdir_LocalPane(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.activePane = localPane
	cmd := sftp.handleMkdir()
	if cmd == nil {
		t.Fatal("expected command for mkdir on local pane")
	}

	msg := cmd()
	transferMsg, ok := msg.(sftpTransferMsg)
	if !ok {
		t.Fatalf("expected sftpTransferMsg, got %T", msg)
	}
	if transferMsg.err == nil {
		t.Error("expected error for mkdir on local pane")
	}
}

func TestSftpModel_fileExistsLocal(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	tmp := t.TempDir()
	testFile := filepath.Join(tmp, "exists.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644) //nolint:gosec

	if !sftp.fileExistsLocal(testFile) {
		t.Error("expected file to exist")
	}
	if sftp.fileExistsLocal(filepath.Join(tmp, "nonexistent.txt")) {
		t.Error("expected file to not exist")
	}
}

func TestSftpModel_fileExistsRemote_NoClient(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	if sftp.fileExistsRemote("/some/path") {
		t.Error("expected false when no client connected")
	}
}

func TestSftpModel_syncPaneSizes(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.syncPaneSizes(100, 30)

	if sftp.width != 100 {
		t.Errorf("width = %d, want 100", sftp.width)
	}
	if sftp.height != 30 {
		t.Errorf("height = %d, want 30", sftp.height)
	}
}

func TestSftpModel_syncPaneSizes_ZeroFallback(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	sftp.syncPaneSizes(0, 0)

	if sftp.width <= 0 {
		t.Error("expected positive width after zero fallback")
	}
	if sftp.height <= 0 {
		t.Error("expected positive height after zero fallback")
	}
}

func TestSftpModel_View(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	v := sftp.View()

	if !v.AltScreen {
		t.Error("expected AltScreen to be true")
	}
}

func TestSftpModel_history_Limit(t *testing.T) {
	m := newTestModel(t, false)
	m.li.CursorDown()

	result, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	sftp := result.(*sftpModel)

	for i := range 60 {
		_, _ = sftp.Update(sftpTransferMsg{
			text:          fmt.Sprintf("transfer %d", i),
			refreshRemote: false,
		})
	}

	if len(sftp.history) > 50 {
		t.Errorf("history length = %d, want <= 50", len(sftp.history))
	}
}

func TestTransferMsgCmd(t *testing.T) {
	cmd := transferMsgCmd("test message", nil, true, false)
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	msg := cmd()
	transferMsg, ok := msg.(sftpTransferMsg)
	if !ok {
		t.Fatalf("expected sftpTransferMsg, got %T", msg)
	}
	if transferMsg.text != "test message" {
		t.Errorf("text = %q, want %q", transferMsg.text, "test message")
	}
	if !transferMsg.refreshLocal {
		t.Error("expected refreshLocal to be true")
	}
	if transferMsg.refreshRemote {
		t.Error("expected refreshRemote to be false")
	}
}

func TestTransferMsgCmd_Error(t *testing.T) {
	testErr := os.ErrNotExist
	cmd := transferMsgCmd("", testErr, false, true)
	msg := cmd()

	transferMsg, ok := msg.(sftpTransferMsg)
	if !ok {
		t.Fatalf("expected sftpTransferMsg, got %T", msg)
	}
	if transferMsg.err != testErr {
		t.Errorf("err = %v, want %v", transferMsg.err, testErr)
	}
}

func TestProgressWriter_Write(t *testing.T) {
	pw := &progressWriter{
		writer: nil,
		name:   "test.txt",
		total:  100,
	}

	var buf []byte
	pw.writer = &testWriter{buf: &buf}

	n, err := pw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != 5 {
		t.Errorf("Write() n = %d, want 5", n)
	}
	if pw.sent != 5 {
		t.Errorf("sent = %d, want 5", pw.sent)
	}
}

type testWriter struct {
	buf *[]byte
}

func (w *testWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
