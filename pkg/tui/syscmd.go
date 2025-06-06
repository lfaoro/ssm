package tui

type SysCmd string

func (s SysCmd) String() string {
	return string(s)
}

const (
	sshCmd  SysCmd = "ssh"
	moshCmd SysCmd = "mosh"
)
