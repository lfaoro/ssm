// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

// SysCmd represents a system command (ssh, mosh, etc.).
type SysCmd string

func (s SysCmd) String() string {
	return string(s)
}

const (
	// SSHCmd is the default backend for direct connections.
	SSHCmd SysCmd = "ssh"
	// MoshCmd selects the mosh backend for direct interactive Enter connections
	// (via Tab or --backend/SSM_BACKEND at startup). Batch exec, Ctrl+r run-command,
	// and SFTP always use ssh regardless.
	MoshCmd SysCmd = "mosh"
)
