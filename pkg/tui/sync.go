// Copyright (c) 2026 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lfaoro/ssm/pkg/syncer"
)

type syncStatus int

const (
	syncIdle syncStatus = iota
	syncRunning
	syncDone
	syncError
)

type provStatus int

const (
	provPending provStatus = iota
	provFetching
	provDone
	provSkipped
	provFailed
)

type provState struct {
	name   string
	label  string
	status provStatus
	count  int
	err    string
}

// SyncModel wraps the base model in a cloud provider sync sub-model.
func SyncModel(base tea.Model) tea.Model {
	previousModel, ok := base.(*Model)
	if !ok {
		return base
	}
	s := spinner.New()
	s.Spinner = spinner.Dot
	provs := []provState{
		{name: "hetzner", label: "Hetzner"},
		{name: "aws", label: "AWS"},
		{name: "gcp", label: "GCP"},
		{name: "azure", label: "Azure"},
	}
	return &syncModel{
		previousModel: base,
		syncer:        syncer.New(),
		providers:     provs,
		status:        syncIdle,
		spinner:       s,
		width:         previousModel.vp.Width() * 2,
		height:        previousModel.vp.Height(),
	}
}

type syncModel struct {
	previousModel tea.Model
	syncer        *syncer.Syncer

	providers []provState
	status    syncStatus
	user      string
	keyPath   string

	spinner     spinner.Model
	width       int
	height      int
	resultLines []string
}

func (m *syncModel) Init() tea.Cmd {
	return nil
}

func (m *syncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEsc, 'q':
			return m.previousModel, nil
		case 's', tea.KeyEnter:
			if m.status == syncIdle || m.status == syncDone || m.status == syncError {
				return m, m.startSync()
			}
		}
	case syncProgressMsg:
		return m.handleProgress(msg)
	case syncCompleteMsg:
		return m.handleComplete(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *syncModel) startSync() tea.Cmd {
	m.status = syncRunning
	m.resultLines = nil
	enabled := make([]string, 0, len(m.providers))
	for i := range m.providers {
		m.providers[i].status = provPending
		m.providers[i].count = 0
		m.providers[i].err = ""
		enabled = append(enabled, m.providers[i].name)
	}

	return func() tea.Msg {
		byProvider, err := m.syncer.Sync(context.Background(), m.user, m.keyPath, enabled)
		if err != nil {
			return syncCompleteMsg{err: err}
		}
		var lines []string
		var total int
		for _, p := range m.providers {
			n := len(byProvider[p.name])
			total += n
			if n > 0 {
				lines = append(lines, fmt.Sprintf("%s: %d servers", p.label, n))
			}
		}
		if total == 0 {
			lines = append(lines, "No servers found (check credentials)")
		}
		progress := make([]syncProgressMsg, 0, len(m.providers))
		for _, p := range m.providers {
			n := len(byProvider[p.name])
			progress = append(progress, syncProgressMsg{
				provider: p.name,
				count:    n,
			})
		}
		return syncCompleteMsg{
			progress: progress,
			summary:  lines,
		}
	}
}

func (m *syncModel) handleProgress(msg syncProgressMsg) (tea.Model, tea.Cmd) {
	for i := range m.providers {
		if m.providers[i].name == msg.provider {
			m.providers[i].status = provFetching
		}
	}
	return m, nil
}

func (m *syncModel) handleComplete(msg syncCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = syncError
		for i := range m.providers {
			m.providers[i].status = provFailed
			m.providers[i].err = msg.err.Error()
		}
		return m, nil
	}
	m.status = syncDone
	for _, p := range msg.progress {
		for i := range m.providers {
			if m.providers[i].name == p.provider {
				m.providers[i].status = provDone
				m.providers[i].count = p.count
			}
		}
	}
	m.resultLines = msg.summary
	return m, nil
}

func (m *syncModel) View() tea.View {
	var b strings.Builder
	b.WriteString(m.renderBar() + "\n\n")
	if m.status == syncRunning {
		b.WriteString(m.spinner.View() + " Syncing cloud providers...\n\n")
	}
	for _, p := range m.providers {
		b.WriteString(m.renderProvider(p) + "\n")
	}
	b.WriteString("\n")
	if len(m.resultLines) > 0 {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Padding(0, 1)
		b.WriteString(style.Render("✓ Sync complete"))
		b.WriteString("\n")
		for _, line := range m.resultLines {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(m.renderFooter())
	v := tea.NewView(b.String())
	v.AltScreen = true
	if pm, ok := m.previousModel.(*Model); ok {
		v.BackgroundColor = parseHexColor(pm.theme.backgroundColor)
	}
	return v
}

func (m *syncModel) renderBar() string {
	pm, ok := m.previousModel.(*Model)
	if !ok {
		return "invalid model"
	}
	title := renderPrimaryBar("Sync Cloud Providers", pm.theme.selectedTitleColor)
	status := renderPrimaryBar("SSM", pm.theme.selectedTitleColor)
	return lipgloss.JoinHorizontal(lipgloss.Top, title, status)
}

func (m *syncModel) renderProvider(p provState) string {
	icon := "○"
	switch p.status {
	case provFetching:
		icon = "⟳"
	case provDone:
		icon = "✓"
	case provSkipped:
		icon = "–"
	case provFailed:
		icon = "✗"
	}
	label := fmt.Sprintf("  %s %s", icon, p.label)
	var info string
	switch {
	case p.status == provFetching:
		info = "  fetching..."
	case p.status == provDone:
		info = fmt.Sprintf("  %d servers", p.count)
	case p.status == provFailed && p.err != "":
		info = "  " + p.err
	case p.status == provSkipped:
		info = "  skipped"
	}
	return label + info
}

func (m *syncModel) renderFooter() string {
	switch m.status {
	case syncIdle:
		return "[Enter] Sync all  [Esc/q] Back"
	case syncDone:
		return "[s] Sync again  [Esc/q] Back"
	case syncError:
		return "[s] Retry  [Esc/q] Back"
	default:
		return "[Esc/q] Back"
	}
}

type syncProgressMsg struct {
	provider string
	count    int
}

type syncCompleteMsg struct {
	progress []syncProgressMsg
	summary  []string
	err      error
}
