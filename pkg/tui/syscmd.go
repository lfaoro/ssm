// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

package tui

// SysCmd represents a system command (ssh, mosh, etc.).
type SysCmd string

func (s SysCmd) String() string {
	return string(s)
}

const (
	sshCmd  SysCmd = "ssh"
	moshCmd SysCmd = "mosh"
)
