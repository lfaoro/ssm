// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

package tui

type (
	// ShowConfigMsg requests the config viewport to be displayed.
	ShowConfigMsg struct{}
	// ReloadConfigMsg requests the SSH config to be re-parsed.
	ReloadConfigMsg struct{}
	// LivenessCheckMsg triggers a ping-based liveness check.
	LivenessCheckMsg struct{}
	// ExitOnConnMsg signals the app to exit after establishing a connection.
	ExitOnConnMsg struct{}
	// SetThemeMsg changes the UI colour theme.
	SetThemeMsg struct {
		Theme string
	}
	tickMsg struct{}
	// AppMsg displays a message in the app status bar.
	AppMsg struct {
		Text string
	}
	// FilterTagMsg sets the host list filter by tag.
	FilterTagMsg struct {
		Arg string
	}
)
