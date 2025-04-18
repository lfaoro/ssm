package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	lp "github.com/charmbracelet/lipgloss/v2"

	"github.com/lfaoro/ssm/pkg/sshconf"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title + i.desc }

func listFrom(config *sshconf.Config) list.Model {
	var li list.Model
	lightDark := lp.LightDark(true)
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetSpacing(0)
	d.Styles.SelectedTitle = lp.NewStyle().
		Border(lp.NormalBorder(), false, false, false, true).
		BorderForeground(lightDark(lp.Color("#F79F3F"), lp.Color("#00bfff"))).
		Foreground(lightDark(lp.Color("#F79F3F"), lp.Color("#00bfff"))).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle.
		Foreground(lightDark(lp.Color("#F79F3F"), lp.Color("#4682b4")))

	editKey := key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "edit config"),
	)
	tabKey := key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch ssh/mosh"),
	)
	enterKey := key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "connect"),
	)

	li = list.New(
		[]list.Item{},
		d,
		0,
		0,
	)
	li.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			editKey,
			tabKey,
			enterKey,
		}
	}

	li.Styles.StatusBar = lp.NewStyle().
		Foreground(lightDark(lp.Color("#A49FA5"), lp.Color("#777777"))).
		Padding(0, 0, 1, 2) //nolint:mnd
	li.Styles.Title = lp.NewStyle().
		Background(lp.Color("#4682b4")).
		Foreground(lp.Color("230")).
		Padding(0, 1)
	li.SetStatusBarItemName("host", "hosts")
	li.Title = fmt.Sprintf("SSH servers (%v)", config.GetPath())
	for _, host := range config.Hosts {
		if host.Name == "*" {
			continue
		}
		fmtDescription := func() string {
			port := func() string {
				_port := host.Options["port"]
				if _port != "" {
					return fmt.Sprintf(":%s", _port)
				}
				return ""
			}
			user := func() string {
				_user := host.Options["user"]
				if _user != "" {
					return _user + "@"
				}
				return ""
			}
			hostname := func() string {
				_host := host.Options["hostname"]
				if _host != "" {
					return _host
				}
				return ""
			}
			out := fmt.Sprintf("%s%s%s", user(), hostname(), port())
			return out
		}()
		newitem := item{
			title: host.Name,
			desc:  fmtDescription,
		}
		li.InsertItem(len(config.Hosts), newitem)
	}
	// ad for segfault
	li.InsertItem(0, item{
		title: "segfault.net",
		desc:  "create free root server",
	})

	return li
}
