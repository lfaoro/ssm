// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

package tui

type (
	ShowConfigMsg    struct{}
	ReloadConfigMsg  struct{}
	LivenessCheckMsg struct{}
	ExitOnConnMsg    struct{}
	SetThemeMsg      struct {
		Theme string
	}
	tickMsg struct{}
	AppMsg  struct {
		Text string
	}
	FilterTagMsg struct {
		Arg string
	}
)
