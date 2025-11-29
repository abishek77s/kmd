package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI state
type model struct {
	choices     []string
	cursor      int
	selected    map[int]struct{}
	order       []int
	finished    bool
	animTick    int
	statusMsg   string
	statusTimer time.Time
}

// Messages for the TUI
type animateMsg struct{}
type statusClearMsg struct{}

func initialModel(cmdCount int) model {
	commands, err := getLastCommands(cmdCount)
	if err != nil {
		commands = []string{fmt.Sprintf("Error reading history: %v", err)}
	}

	return model{
		choices:     commands,
		selected:    make(map[int]struct{}),
		order:       make([]int, 0),
		animTick:    0,
		statusTimer: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.clearStatusCmd(),
	)
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(time.Time) tea.Msg {
		return animateMsg{}
	})
}

func (m model) clearStatusCmd() tea.Cmd {
	return tea.Tick(time.Second*3, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case animateMsg:
		m.animTick = (m.animTick + 1) % len(spinnerFrames)
		return m, m.tickCmd()

	case statusClearMsg:
		if time.Since(m.statusTimer) > time.Second*3 {
			m.statusMsg = ""
		}
		return m, nil

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.MouseWheelDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case tea.MouseRight:
			m.toggleSelection()
			return m, m.clearStatusCmd()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.toggleSelection()
			return m, m.clearStatusCmd()

		case "g":
			if len(m.selected) > 0 {
				m.finished = true
				return m, tea.Sequence(
					func() tea.Msg { return generateMakefile(m.choices, m.order) },
					tea.Quit,
				)
			}
			m.statusMsg = "No commands selected!"
			m.statusTimer = time.Now()
			return m, m.clearStatusCmd()

		case "r":
			return m.refresh()
		}
	}

	return m, nil
}

func (m *model) toggleSelection() {
	if _, ok := m.selected[m.cursor]; ok {
		delete(m.selected, m.cursor)
		for i, idx := range m.order {
			if idx == m.cursor {
				m.order = append(m.order[:i], m.order[i+1:]...)
				break
			}
		}
		m.statusMsg = "Command deselected"
	} else {
		m.selected[m.cursor] = struct{}{}
		m.order = append(m.order, m.cursor)
		m.statusMsg = fmt.Sprintf("Selected (%d total)", len(m.selected))
	}
	m.statusTimer = time.Now()
}

func (m model) refresh() (tea.Model, tea.Cmd) {
	cmdCount := len(m.choices)
	commands, err := getLastCommands(cmdCount)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
	} else {
		m.choices = commands
		m.cursor = 0
		m.selected = make(map[int]struct{})
		m.order = make([]int, 0)
		m.statusMsg = "History refreshed"
	}
	m.statusTimer = time.Now()
	return m, m.clearStatusCmd()
}

func (m model) View() string {
	if m.finished {
		return statusStyle.Render("✨ Makefile generated! Check ./Makefile") + "\n"
	}

	var s strings.Builder

	// Title
	title := fmt.Sprintf(" kmd %s", spinnerFrames[m.animTick])
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")

	// Command list
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}

		checkbox := " "
		order := ""
		if _, ok := m.selected[i]; ok {
			checkbox = selectedStyle.Render("✓")
			for orderIdx, selectedIdx := range m.order {
				if selectedIdx == i {
					order = orderStyle.Render(fmt.Sprintf(" (%d)", orderIdx+1))
					break
				}
			}
		}

		displayCmd := choice
		if len(displayCmd) > 60 {
			displayCmd = displayCmd[:57] + "..."
		}

		line := fmt.Sprintf("%s [%s] %s%s", cursor, checkbox, commandStyle.Render(displayCmd), order)
		s.WriteString(line + "\n")
	}

	s.WriteString("\n")

	// Status message
	if m.statusMsg != "" {
		if strings.Contains(m.statusMsg, "Error") {
			s.WriteString(errorStyle.Render(m.statusMsg))
		} else {
			s.WriteString(statusStyle.Render(m.statusMsg))
		}
		s.WriteString("\n")
	}

	// Controls
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D3D3D3"))
	controls := "↑/↓: Navigate  •  Space/Enter: Toggle  •  g: Generate  •  r: Refresh  •  q: Quit"
	s.WriteString(controlsStyle.Render(controls))

	return s.String()
}
