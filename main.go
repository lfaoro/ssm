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
	"github.com/lfaoro/ssm/pkg/syncer"
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
		UsageText:              "ssm [--options] [tag]\nexample: ssm --show --exit vpn\nexample: ssm -se vpn\nexample: ssm exec prod 'uptime' --delay 200ms\nexample (legacy): ssm dev -r 'whoami && pwd'",
		ArgsUsage:              "[tag]",
		Description:            "SSM is an open-source terminal UI that sits on top of your existing SSH config to simplify and automate connectivity, data transfer, organization and host discovery.",

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

		// Root flags split for subcommand UX:
		// - Local: true  → TUI-only or legacy shims. Not inherited by subcommands
		//   (ssm exec, ssm sync, ...) and do not appear in their "GLOBAL OPTIONS".
		// - (no Local)   → truly cross-cutting. Still visible and usable under
		//   subcommands. We keep only --debug and --config in this category.
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "always show config params",
				Value:   false,
				Sources: cli.EnvVars("SSM_SHOW"),
				Local:   true,
			},
			&cli.BoolFlag{
				Name:    "exit",
				Aliases: []string{"e"},
				Usage:   "exit after connection",
				Value:   false,
				Sources: cli.EnvVars("SSM_EXIT"),
				Local:   true,
			},
			&cli.BoolFlag{
				Name:    "order",
				Aliases: []string{"o"},
				Usage:   "show hosts with a tag first",
				Value:   false,
				Sources: cli.EnvVars("SSM_ORDER"),
				Local:   true,
			},
			&cli.BoolFlag{
				Name:    "ping",
				Aliases: []string{"p"},
				Usage:   "connect to all hosts and show response time",
				Value:   false,
				Sources: cli.EnvVars("SSM_PING"),
				Local:   true,
			},
			&cli.StringFlag{
				Name:      "config",
				TakesFile: true,
				Aliases:   []string{"c"},
				Usage:     "custom ssh config file path",
				Sources:   cli.EnvVars("SSM_SSH_CONFIG_PATH"),
				// intentionally not Local — useful for exec/sync too
			},
			&cli.StringFlag{
				Name:        "theme",
				TakesFile:   false,
				Aliases:     []string{"t"},
				Usage:       "define a color theme",
				DefaultText: "sky|matrix",
				Value:       "sky",
				Sources:     cli.EnvVars("SSM_THEME"),
				Local:       true,
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug mode with verbose logging",
				Value:   false,
				Sources: cli.EnvVars("SSM_DEBUG"),
				// intentionally not Local — useful everywhere
			},
			&cli.StringFlag{
				Name:    "command",
				Aliases: []string{"r"},
				Usage:   "run command on all (or tag-filtered) hosts and exit (non-interactive) [deprecated: use `ssm exec` instead]",
				Sources: cli.EnvVars("SSM_COMMAND"),
				Local:   true,
				// Note: we do NOT set Hidden: true. It will still appear (with the
				// deprecation note) in the root `ssm --help`, just not under exec/sync.
			},
		},

		Commands: []*cli.Command{
			generateCmd,
			testCmd,
			syncCmd,
			execCmd,
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

	// Load config early (needed for both TUI and --command batch paths).
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

	// --command (or -r) is a non-interactive batch path (legacy shim).
	// It calls the 6-arg form with zeros so defaults + modest auto jitter apply.
	if command := cmd.String("command"); command != "" {
		return tui.RunBatchRemoteCommands(config, filterTag, command, 0, 0, 0)
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("not an interactive terminal :(")
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
	if filterTag != "" {
		p.Send(tui.FilterTagMsg{
			Arg: "#" + filterTag,
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

var syncCmd = &cli.Command{
	Name:      "sync",
	Usage:     "Sync servers from cloud providers into SSH config",
	ArgsUsage: "[hetzner aws gcp azure]",
	Description: `Discover running servers from cloud providers and write them to ~/.ssh/config.d/50-ssm-{provider}.

Credentials are read from environment variables:
  Hetzner  HCLOUD_TOKEN
  AWS      Standard SDK chain (AWS_PROFILE, AWS_ACCESS_KEY_ID, IAM role)
  GCP      GCP_PROJECT + Application Default Credentials
  Azure    AZURE_SUBSCRIPTION_ID + Azure SDK auth (env, CLI, managed identity)

Each provider gets its own file under ~/.ssh/config.d/. The Include config.d/*
directive is automatically added to ~/.ssh/config if missing.`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"n"},
			Usage:   "preview generated config without writing",
			Sources: cli.EnvVars("SSM_SYNC_DRY_RUN"),
		},
		&cli.StringFlag{
			Name:  "user",
			Usage: "default SSH user for all synced hosts",
		},
		&cli.StringFlag{
			Name:  "key",
			Usage: "default IdentityFile path for all synced hosts",
		},
	},
	Action: syncAction,
}

var syncAction = func(_ context.Context, cmd *cli.Command) error {
	s := syncer.New()
	providers := cmd.Args().Slice()

	if cmd.Bool("dry-run") {
		content, err := s.DryRun(context.Background(), cmd.String("user"), cmd.String("key"), providers)
		if err != nil {
			return err
		}
		fmt.Println(content)
		return nil
	}

	byProvider, err := s.Sync(context.Background(), cmd.String("user"), cmd.String("key"), providers)
	if err != nil {
		return err
	}

	if len(byProvider) == 0 {
		fmt.Println("ssm: no servers synced (check credentials and provider names)")
		return nil
	}

	var total int
	var pathList []string
	for _, p := range []string{"hetzner", "aws", "gcp", "azure"} {
		if n := len(byProvider[p]); n > 0 {
			total += n
			fmt.Printf("ssm: synced %d servers: %s=%d\n", total, p, n)
			pathList = append(pathList, "  "+s.Path(p))
		}
	}
	fmt.Println("ssm: config written:")
	for _, p := range pathList {
		fmt.Println(p)
	}
	return nil
}

var execCmd = &cli.Command{
	Name:      "exec",
	Aliases:   []string{"e"},
	Usage:     "Run a command on all (or tag-filtered) hosts non-interactively",
	ArgsUsage: "[tag] 'command'",
	Description: `Equivalent to the legacy -r/--command flag but with full control
over pacing and concurrency. Commands with spaces must be quoted.

Examples:
  ssm exec 'uptime'
  ssm exec prod 'uptime && whoami'
  ssm exec web --delay 150ms --threads 4 'nginx -t'`,
	Flags: []cli.Flag{
		&cli.DurationFlag{
			Name:    "delay",
			Usage:   "fixed delay between starting command execution on each host (e.g. 100ms, 500ms)",
			Sources: cli.EnvVars("SSM_EXEC_DELAY"),
		},
		&cli.IntFlag{
			Name:    "threads",
			Aliases: []string{"t", "j", "jobs"},
			Usage:   "maximum number of concurrent hosts (default: auto based on CPU)",
			Sources: cli.EnvVars("SSM_EXEC_THREADS"),
		},
		&cli.DurationFlag{
			Name:    "jitter-max",
			Usage:   "maximum random jitter added to each inter-host delay (default: modest auto jitter)",
			Sources: cli.EnvVars("SSM_EXEC_JITTER_MAX"),
		},
	},
	Action: execAction,
}

var execAction = func(_ context.Context, cmd *cli.Command) error {
	// Config loading duplicated from mainCmd for subcommand independence.
	// This keeps the change focused; a later refactor can extract a helper.
	var err error
	cfg := sshconf.New()
	if cmd.Bool("order") { // unlikely to be set on subcommand, but harmless
		cfg.SetOrder(sshconf.TagOrder)
	}
	configFlag := cmd.String("config")
	if configFlag != "" {
		err = cfg.ParsePath(configFlag)
	} else {
		err = cfg.Parse()
	}
	if err != nil {
		return err
	}

	tagFilter := "" // subcommand can parse its own positionals
	args := cmd.Args().Slice()
	command := ""
	switch len(args) {
	case 0:
		return errors.New("ssm exec: command argument is required (quote it if it contains spaces)")
	case 1:
		command = args[0]
	default:
		// first arg is tag filter, last arg is the command
		tagFilter = args[0]
		command = args[len(args)-1]
	}
	if command == "" {
		return errors.New("ssm exec: command argument is required")
	}

	delay := cmd.Duration("delay")
	threads := cmd.Int("threads")
	jitterMax := cmd.Duration("jitter-max")

	return tui.RunBatchRemoteCommands(cfg, tagFilter, command, delay, threads, jitterMax)
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
