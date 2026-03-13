package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Attach key.Binding
	Peek   key.Binding
	New    key.Binding
	Delete key.Binding
	Stop   key.Binding
	Resume key.Binding
	Help   key.Binding
	Quit   key.Binding
	Escape key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k/up", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j/down", "down"),
	),
	Attach: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "attach"),
	),
	Peek: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "peek"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Stop: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "stop"),
	),
	Resume: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "resume"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

// statusBarHelp returns the formatted keybinding hints for the status bar.
func statusBarHelp() string {
	bindings := []key.Binding{
		keys.Attach, keys.Delete, keys.Stop,
		keys.Resume, keys.New, keys.Quit,
	}

	var s string
	for i, b := range bindings {
		if i > 0 {
			s += "  "
		}
		h := b.Help()
		s += helpKeyStyle.Render(h.Key) + ":" + helpDescStyle.Render(h.Desc)
	}
	return s
}
