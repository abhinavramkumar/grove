package app

import (
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// appTestModel wraps AppModel for teatest, skipping Init's async commands
// (reconcileAndLoad needs tmux, tickCmd runs forever).
type appTestModel struct {
	app AppModel
}

func (m appTestModel) Init() tea.Cmd                           { return nil }
func (m appTestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := m.app.Update(msg)
	m.app = result.(AppModel)
	return m, cmd
}
func (m appTestModel) View() string { return stabilizeView(m.app.View()) }

func newIntegrationModel(t *testing.T) *teatest.TestModel {
	t.Helper()
	sessions := makeTestSessions()
	app := makeTestAppModel(t, sessions)
	wrapper := appTestModel{app: app}
	tm := teatest.NewTestModel(t, wrapper, teatest.WithInitialTermSize(80, 24))
	time.Sleep(100 * time.Millisecond) // let initial render complete
	return tm
}

func TestIntegration_ListToHelp(t *testing.T) {
	tm := newIntegrationModel(t)

	// Open help overlay.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	time.Sleep(100 * time.Millisecond)

	// Close help with esc (q is not handled in help view).
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})
	time.Sleep(100 * time.Millisecond)

	// Quit from list view.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestIntegration_ListToThemePicker(t *testing.T) {
	withTheme(t, TokyoNight)
	tm := newIntegrationModel(t)

	// Open theme picker.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	time.Sleep(100 * time.Millisecond)

	// Close with esc.
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})
	time.Sleep(100 * time.Millisecond)

	// Quit.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestIntegration_CursorNavigation(t *testing.T) {
	tm := newIntegrationModel(t)

	// Move cursor: j, j, k
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	time.Sleep(50 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestIntegration_FilterFlow(t *testing.T) {
	tm := newIntegrationModel(t)

	// Start filter with ctrl+f.
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlF})
	time.Sleep(100 * time.Millisecond)

	// Type "auth".
	for _, r := range "auth" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		time.Sleep(30 * time.Millisecond)
	}

	// Commit filter.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Quit.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestIntegration_HelpKeys(t *testing.T) {
	tm := newIntegrationModel(t)

	// Open help.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	time.Sleep(100 * time.Millisecond)

	// Close help with esc.
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})
	time.Sleep(100 * time.Millisecond)

	// Quit.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(5*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	teatest.RequireEqualOutput(t, out)
}
