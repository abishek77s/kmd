package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string         // command history items
	cursor   int              // which item our cursor is pointing at
	selected map[int]struct{} // which items are selected
	order    []int            // order in which items were selected
	finished bool             // whether we're done selecting
}

func init() {
	// Ensure proper history settings are configured on startup
	setupHistorySettings()
}

func setupHistorySettings() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	shell := os.Getenv("SHELL")

	// Force history to be written immediately for current session
	if strings.Contains(shell, "bash") {
		exec.Command("bash", "-c", "history -a").Run()

		// Setup persistent bash history configuration
		bashrcPath := filepath.Join(usr.HomeDir, ".bashrc")
		return ensureHistoryConfig(bashrcPath, []string{
			"# Auto-save history after each command (added by cmd-makefile)",
			"export PROMPT_COMMAND='history -a'",
			"export HISTSIZE=10000",
			"export HISTFILESIZE=20000",
			"shopt -s histappend",
		})
	} else if strings.Contains(shell, "zsh") {
		// Setup persistent zsh history configuration
		zshrcPath := filepath.Join(usr.HomeDir, ".zshrc")
		return ensureHistoryConfig(zshrcPath, []string{
			"# Auto-save history settings (added by cmd-makefile)",
			"setopt SHARE_HISTORY",
			"setopt HIST_SAVE_NO_DUPS",
			"setopt HIST_IGNORE_ALL_DUPS",
			"setopt HIST_FIND_NO_DUPS",
			"setopt HIST_IGNORE_SPACE",
			"setopt APPEND_HISTORY",
			"setopt INC_APPEND_HISTORY",
			"export HISTSIZE=10000",
			"export SAVEHIST=10000",
		})
	}

	return nil
}

func ensureHistoryConfig(configFile string, configLines []string) error {
	// Read existing config file
	content, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	existingContent := string(content)
	marker := "# cmd-makefile history config"

	// Check if our config is already present
	if strings.Contains(existingContent, marker) {
		return nil // Already configured
	}

	// Append our configuration
	file, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add a newline if file doesn't end with one
	if len(existingContent) > 0 && !strings.HasSuffix(existingContent, "\n") {
		file.WriteString("\n")
	}

	// Add our configuration block
	file.WriteString("\n" + marker + "\n")
	for _, line := range configLines {
		file.WriteString(line + "\n")
	}
	file.WriteString("# end cmd-makefile config\n")

	fmt.Printf("✓ Added history configuration to %s\n", configFile)
	fmt.Printf("  Run 'source %s' or start a new terminal session for changes to take effect.\n", configFile)

	return nil
}

func getLastCommands(count int) ([]string, error) {
	// Force immediate history save
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") {
		exec.Command("bash", "-c", "history -a").Run()
	}

	// First try to get current session history
	commands, err := getCurrentSessionHistory(count)
	if err == nil && len(commands) > 0 {
		return commands, nil
	}

	// Fallback to reading history files
	return getHistoryFromFiles(count)
}

func getCurrentSessionHistory(count int) ([]string, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	var cmd *exec.Cmd

	// Use different approaches based on shell
	if strings.Contains(shell, "bash") {
		// For bash: reload history and get recent commands
		historyCmd := fmt.Sprintf(`
			# Force reload history from file
			history -r 2>/dev/null || true
			# Get last commands, but more than requested to account for filtering
			history %d 2>/dev/null | tail -n %d
		`, count+20, count+10)
		cmd = exec.Command("bash", "-c", historyCmd)
	} else if strings.Contains(shell, "zsh") {
		// For zsh: get from current session with fc
		historyCmd := fmt.Sprintf(`
			fc -l -%d 2>/dev/null | tail -n %d
		`, count+20, count+10)
		cmd = exec.Command("zsh", "-c", historyCmd)
	} else {
		// Generic fallback
		cmd = exec.Command(shell, "-c", fmt.Sprintf("history %d 2>/dev/null || fc -l -%d 2>/dev/null", count+10, count+10))
	}

	// Set environment to inherit history settings
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return getHistoryFromParentShell(count)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var commands []string
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var command string

		// Parse different history formats
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Try to find number at start: "  123  command here" or "123  command here"
			if _, err := strconv.Atoi(parts[0]); err == nil {
				command = strings.Join(parts[1:], " ")
			} else {
				command = line // Use whole line if no number prefix
			}
		} else {
			command = line
		}

		if command != "" && !seen[command] {
			// Filter out this program and common history commands
			if !strings.Contains(command, "go run main.go") &&
				!strings.Contains(command, "./cmd-makefile") &&
				!strings.Contains(command, "cmd-makefile") &&
				!strings.HasPrefix(command, "history") &&
				!strings.HasPrefix(command, "fc -l") &&
				len(command) > 2 { // Skip very short commands

				commands = append(commands, command)
				seen[command] = true
			}
		}
	}

	// Return the most recent commands
	if len(commands) > count {
		commands = commands[len(commands)-count:]
	}

	return commands, nil
}

func getHistoryFromParentShell(count int) ([]string, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	shell := os.Getenv("SHELL")
	var historyFile string

	if strings.Contains(shell, "bash") {
		historyFile = filepath.Join(usr.HomeDir, ".bash_history")
		// Try to force write current session to history file
		exec.Command("bash", "-c", "history -a").Run()
	} else if strings.Contains(shell, "zsh") {
		historyFile = filepath.Join(usr.HomeDir, ".zsh_history")
	} else {
		historyFile = filepath.Join(usr.HomeDir, ".history")
	}

	return readHistoryFile(historyFile, count)
}

func readHistoryFile(filename string, count int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	seen := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle zsh extended history format: ": 1234567890:0;command"
		if strings.HasPrefix(line, ":") && strings.Contains(line, ";") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				line = strings.TrimSpace(parts[1])
			}
		}

		// Skip duplicates and filter commands
		if !seen[line] && len(line) > 2 {
			if !strings.Contains(line, "go run main.go") &&
				!strings.Contains(line, "./cmd-makefile") &&
				!strings.Contains(line, "cmd-makefile") &&
				!strings.HasPrefix(line, "history") &&
				!strings.HasPrefix(line, "fc -l") {

				allLines = append(allLines, line)
				seen[line] = true
			}
		}
	}

	if len(allLines) == 0 {
		return nil, fmt.Errorf("no commands found in history")
	}

	// Get the last 'count' commands
	start := len(allLines) - count
	if start < 0 {
		start = 0
	}

	return allLines[start:], nil
}

func getHistoryFromFiles(count int) ([]string, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// Try different history file locations
	historyFiles := []string{
		filepath.Join(usr.HomeDir, ".bash_history"),
		filepath.Join(usr.HomeDir, ".zsh_history"),
		filepath.Join(usr.HomeDir, ".history"),
	}

	for _, histFile := range historyFiles {
		if commands, err := readHistoryFile(histFile, count); err == nil && len(commands) > 0 {
			return commands, nil
		}
	}

	return []string{"echo 'No command history found - try running some commands first'"}, nil
}

func initialModel() model {
	commands, err := getLastCommands(5)
	if err != nil {
		commands = []string{fmt.Sprintf("Error reading history: %v", err)}
	}

	return model{
		choices:  commands,
		selected: make(map[int]struct{}),
		order:    make([]int, 0),
		finished: false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			_, ok := m.selected[m.cursor]
			if ok {
				// Remove from selection and order
				delete(m.selected, m.cursor)
				for i, idx := range m.order {
					if idx == m.cursor {
						m.order = append(m.order[:i], m.order[i+1:]...)
						break
					}
				}
			} else {
				// Add to selection and order
				m.selected[m.cursor] = struct{}{}
				m.order = append(m.order, m.cursor)
			}
		case "g":
			// Generate makefile when 'g' is pressed
			if len(m.selected) > 0 {
				m.finished = true
				return m, tea.Sequence(
					func() tea.Msg { return generateMakefile(m.choices, m.order) },
					tea.Quit,
				)
			}
		case "r":
			// Refresh command history
			commands, err := getLastCommands(5)
			if err != nil {
				commands = []string{fmt.Sprintf("Error reading history: %v", err)}
			}
			m.choices = commands
			m.cursor = 0
			m.selected = make(map[int]struct{})
			m.order = make([]int, 0)
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.finished {
		return "Makefile generated! Check ./Makefile\n"
	}

	s := "Last 5 Terminal Commands (Oldest → Newest):\n"
	s += "Select commands in the order you want them in the Makefile\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		order := ""
		if _, ok := m.selected[i]; ok {
			checked = "✓"
			// Find the order number
			for orderIdx, selectedIdx := range m.order {
				if selectedIdx == i {
					order = fmt.Sprintf(" (%d)", orderIdx+1)
					break
				}
			}
		}

		s += fmt.Sprintf("%s [%s] %s%s\n", cursor, checked, choice, order)
	}

	s += "\nControls:\n"
	s += "• ↑/↓ or k/j: Navigate\n"
	s += "• Space/Enter: Toggle selection\n"
	s += "• g: Generate Makefile\n"
	s += "• r: Refresh history\n"
	s += "• q: Quit\n"

	return s
}

type makefileMsg struct {
	success bool
	error   error
}

func generateMakefile(choices []string, order []int) makefileMsg {
	if len(order) == 0 {
		return makefileMsg{false, fmt.Errorf("no commands selected")}
	}

	content := "# Generated Makefile from command history\n"
	content += "# Run with: make all\n\n"
	content += ".PHONY: all clean\n\n"
	content += "all:"

	// Add target dependencies
	for i := range order {
		content += fmt.Sprintf(" step%d", i+1)
	}
	content += "\n\n"

	// Add individual steps
	for i, cmdIdx := range order {
		stepName := fmt.Sprintf("step%d", i+1)
		command := choices[cmdIdx]

		content += fmt.Sprintf("%s:\n", stepName)
		content += fmt.Sprintf("\t@echo \"Step %d: %s\"\n", i+1, command)
		content += fmt.Sprintf("\t%s\n\n", command)
	}

	// Add clean target
	content += "clean:\n"
	content += "\t@echo \"Cleaning up...\"\n"
	content += "\t# Add cleanup commands here\n"

	err := os.WriteFile("Makefile", []byte(content), 0644)
	return makefileMsg{success: err == nil, error: err}
}

func main() {
	fmt.Println("🔧 Command History to Makefile Generator")
	fmt.Println("Setting up history configuration...")

	if err := setupHistorySettings(); err != nil {
		fmt.Printf("Warning: Could not setup history configuration: %v\n", err)
		fmt.Println("You may need to manually add 'export PROMPT_COMMAND=\"history -a\"' to your shell config.")
	}

	fmt.Println()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
