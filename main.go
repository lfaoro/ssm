// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: MIT

// Package main implements the ssm (Secure Shell Manager) CLI entry point.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/google/go-github/v69/github"
	"github.com/lfaoro/ssm/pkg/sshconf"
	"github.com/lfaoro/ssm/pkg/tui"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var BuildVersion = "0.0.0-dev"
var BuildDate = "unset"
var BuildSHA = "unset"

// cli arguments
var (
	filterTag string
)

func main() {
	appcmd := &cli.Command{
		Name: "ssm",
		Authors: []any{
			&mail.Address{
				Name:    "Leonardo Faoro",
				Address: "me@leonardofaoro.com",
			},
		},
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Suggest:                true,
		Copyright:              "(c) Leonardo Faoro & authors",
		Usage:                  "Secure Shell Manager",
		UsageText:              "ssm [--options] [tag]\nexample: ssm --show --exit vpn\nexample: ssm -se vpn",
		ArgsUsage:              "[tag]",
		Description:            "ssm is a connection manager designed to help organize servers, connect, filter, tag, and much more from a simple terminal interface. It works on top of installed command-line programs and does not require any setup on remote systems.",

		Version: BuildVersion,
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"Build version": BuildVersion,
				"Build date":    BuildDate,
				"Build sha":     BuildSHA,
			}
		},

		Before: func(c context.Context, _ *cli.Command) (context.Context, error) {
			return c, nil
		},

		Action: mainCmd,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "tag",
				UsageText:   "comma separated arguments for filtering #tag: hosts",
				Destination: &filterTag,
			},
		},

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "always show config params",
				Value:   false,
				Sources: cli.EnvVars("SSM_SHOW"),
			},
			&cli.BoolFlag{
				Name:    "exit",
				Aliases: []string{"e"},
				Usage:   "exit after connection",
				Value:   false,
				Sources: cli.EnvVars("SSM_EXIT"),
			},
			&cli.BoolFlag{
				Name:    "order",
				Aliases: []string{"o"},
				Usage:   "show hosts with a tag first",
				Value:   false,
				Sources: cli.EnvVars("SSM_ORDER"),
			},
			&cli.StringFlag{
				Name:      "config",
				TakesFile: true,
				Aliases:   []string{"c"},
				Usage:     "custom ssh config file path",
				Sources:   cli.EnvVars("SSM_SSH_CONFIG_PATH"),
			},
			&cli.StringFlag{
				Name:        "theme",
				TakesFile:   false,
				Aliases:     []string{"t"},
				Usage:       "define a color theme",
				DefaultText: "sky|matrix",
				Value:       "sky",
				Sources:     cli.EnvVars("SSM_THEME"),
			},
			&cli.BoolFlag{
				Name:    "ping",
				Aliases: []string{"p"},
				Usage:   "ping all hosts and show liveness",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug mode with verbose logging",
				Value:   false,
				Sources: cli.EnvVars("SSM_DEBUG"),
			},
		},

		Commands: []*cli.Command{
			generateCmd,
			testCmd,
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
		return errors.New("not an interactive terminal :(")
	}

	var err error
	var config = sshconf.New()
	if cmd.Bool("order") {
		config.SetOrder(sshconf.TagOrder)
	}
	configFlag := cmd.String("config")
	if configFlag != "" {
		err = config.ParsePath(configFlag)
		if err != nil {
			return err
		}
	} else {
		err = config.Parse()
		if err != nil {
			return err
		}
	}

	m := tui.NewModel(config, debug)
	p := tea.NewProgram(
		m,
		tea.WithOutput(os.Stderr))
	wg := sync.WaitGroup{}
	var shutdown atomic.Bool
	wg.Go(func() {
		final, err := p.Run()
		shutdown.Store(true)
		if err != nil {
			e := fmt.Errorf("failed to run %v: %w", cmd.Name, err)
			fmt.Println(e)
			os.Exit(1)
		}
		m, ok := final.(*tui.Model)
		if !ok {
			fmt.Println("you found bug#1: open an issue")
			os.Exit(1)
		}
		if m.ExitOnCmd && m.ExitHost != "" {
			sshPath, err := exec.LookPath(m.Cmd.String())
			if err != nil {
				fmt.Printf("can't find `%s` cmd in your path: %v\n", m.Cmd, err)
				os.Exit(1)
			}
			err = syscall.Exec(sshPath, []string{"ssh", "-F", config.GetPath(), "--", m.ExitHost}, os.Environ()) //nolint:gosec
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	})

	if filterTag != "" {
		p.Send(tui.FilterTagMsg{
			Arg: "#" + filterTag,
		})
	}
	if cmd.Bool("exit") {
		p.Send(tui.ExitOnConnMsg{})
	}
	if cmd.Bool("show") {
		p.Send(tui.ShowConfigMsg{})
	}
	theme := cmd.String("theme")
	if theme != "" {
		p.Send(tui.SetThemeMsg{
			Theme: theme,
		})
	}
	if cmd.Bool("ping") {
		p.Send(tui.LivenessCheckMsg{})
	}

	// inform user when new version is available
	wg.Go(func() {
		tag, err := latestTag()
		if err != nil {
			if cmd.Bool("debug") {
				if !shutdown.Load() {
					p.Send(tui.AppMsg{Text: fmt.Sprintf("%s", err)})
				}
			}
			return
		}
		if tag != cmd.Version && cmd.Version != "0.0.0-dev" {
			if !shutdown.Load() {
				msg := fmt.Sprintf("%s: new version %s is available", cmd.Version, tag)
				p.Send(tui.AppMsg{Text: msg})
			}
		}
	})

	wg.Wait()
	return nil
}

var testCmd = &cli.Command{
	Name:   "test",
	Action: testAction,
	Hidden: true,
}
var testAction = func(_ context.Context, _ *cli.Command) error {
	return nil
}

var generateCmd = &cli.Command{
	Name:    "generate",
	Aliases: []string{"gen"},
	Action:  generateAction,
	Hidden:  true,
}
var generateAction = func(_ context.Context, _ *cli.Command) error {
	return nil
}

func latestTag() (string, error) {
	client := github.NewClient(nil)
	owner := "lfaoro"
	repo := "ssm"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return "", fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		return "", errors.New("no tags found in the repository")
	}

	return *tags[0].Name, nil
}
