package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/lfaoro/ssm/pkg/tui"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var BuildVersion = "v0.0.1-dev"
var BuildDate = "unset"
var BuildSHA = "unset"

func main() {
	appcmd := &cli.Command{
		Authors: []any{
			map[string]string{
				"name":  "Leonardo Faoro",
				"email": "ssm@leonardofaoro.com",
			},
		},
		Name:                   "ssm",
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Suggest:                true,

		Version: fmt.Sprintf("%s\nbuild date: %s\nbuild SHA: %s", BuildVersion, BuildDate, BuildSHA),
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Build version": BuildVersion,
				"Build date":    BuildDate,
				"Build sha":     BuildSHA,
			}
		},
		Usage: "Secure Shell Manager",

		Before: func(c context.Context, cmd *cli.Command) (context.Context, error) {
			_ = cmd
			return c, nil
		},

		Action: mainCmd,

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug mode with verbose logging",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "custom config file path",
			},
		},

		Commands: []*cli.Command{
			subCmd,
		},
	}

	err := appcmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainCmd(_ context.Context, cmd *cli.Command) error {
	debug := cmd.Bool("debug")
	if debug {
		for k, v := range cmd.ExtraInfo() {
			fmt.Println(k, v)
		}
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("not an interactive terminal :(")
	}

	var err error
	var config *sshconf.Config
	configFlag := cmd.String("config")
	if configFlag != "" {
		config, err = sshconf.ParsePath(configFlag)
		if err != nil {
			return err
		}
	} else {
		config, err = sshconf.Parse()
		if err != nil {
			return err
		}
	}

	m := tui.NewModel(config, debug)
	_, err = tea.NewProgram(
		m,
		tea.WithOutput(os.Stderr)).
		Run()
	if err != nil {
		return fmt.Errorf("failed to run %v interface: %w", cmd.Name, err)
	}

	return nil
}

var subCmd = &cli.Command{
	Name:   "test",
	Usage:  "test command",
	Action: subAction,
	Hidden: true,
}

var subAction = func(_ context.Context, _ *cli.Command) error {
	fmt.Println("PID: ", os.Getpid())
	fmt.Println("testing")
	c, err := sshconf.Parse()
	if err != nil {
		return err
	}
	err = c.Watch()
	if err != nil {
		return err
	}

	fmt.Println("done")
	return nil
}
