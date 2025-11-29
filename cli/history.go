package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// getLastCommands retrieves the last N commands from shell history
func getLastCommands(count int) ([]string, error) {
	shellConfig, err := detectShell()
	if err != nil {
		return nil, err
	}

	// Try to flush current session history first
	flushCurrentHistory(shellConfig.Type)

	// Try reading from current session
	if commands, err := readCurrentSessionHistory(shellConfig.Type, count); err == nil && len(commands) > 0 {
		return commands, nil
	}

	// Fall back to history file
	return readHistoryFile(shellConfig.HistoryFile, shellConfig.Type, count)
}

// flushCurrentHistory attempts to flush current session history to file
func flushCurrentHistory(shellType string) {
	switch shellType {
	case ShellBash:
		exec.Command("bash", "-c", "history -a").Run()
	case ShellZsh:
		exec.Command("zsh", "-c", "fc -W").Run()
	}
}

// readCurrentSessionHistory reads history from current shell session
func readCurrentSessionHistory(shellType string, count int) ([]string, error) {
	var cmd *exec.Cmd

	switch shellType {
	case ShellBash:
		histCmd := fmt.Sprintf("history -r 2>/dev/null || true; history %d 2>/dev/null | tail -n %d", count+20, count+10)
		cmd = exec.Command("bash", "-c", histCmd)

	case ShellZsh:
		histCmd := fmt.Sprintf("fc -l -%d 2>/dev/null | tail -n %d", count+20, count+10)
		cmd = exec.Command("zsh", "-c", histCmd)

	case ShellFish:
		histCmd := fmt.Sprintf("history --max=%d 2>/dev/null", count+10)
		cmd = exec.Command("fish", "-c", histCmd)

	default:
		return nil, fmt.Errorf("unsupported shell type: %s", shellType)
	}

	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil, fmt.Errorf("could not read session history")
	}

	return parseHistoryOutput(string(output), count)
}

// readHistoryFile reads and parses commands from history file
func readHistoryFile(filename, shellType string, count int) ([]string, error) {
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

		// Parse based on shell type
		command := parseHistoryLine(line, shellType)
		if command == "" {
			continue
		}

		if !seen[command] && isValidCommand(command) {
			allLines = append(allLines, command)
			seen[command] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(allLines) == 0 {
		return nil, fmt.Errorf("no commands found in history")
	}

	// Return last N commands
	start := len(allLines) - count
	if start < 0 {
		start = 0
	}

	return allLines[start:], nil
}

// parseHistoryLine parses a history line based on shell format
func parseHistoryLine(line, shellType string) string {
	switch shellType {
	case ShellZsh:
		// Zsh format: ": timestamp:0;command"
		if strings.HasPrefix(line, ":") && strings.Contains(line, ";") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}

	case ShellFish:
		// Fish format: "- cmd: command\n  when: timestamp"
		if strings.HasPrefix(line, "- cmd:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "- cmd:"))
		}
		return "" // Skip non-command lines
	}

	return line
}

// parseHistoryOutput parses the output from history command
func parseHistoryOutput(output string, count int) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var commands []string
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove line number prefix (e.g., "  123  command" -> "command")
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		var command string
		if _, err := strconv.Atoi(parts[0]); err == nil {
			command = strings.Join(parts[1:], " ")
		} else {
			command = line
		}

		if !seen[command] && isValidCommand(command) {
			commands = append(commands, command)
			seen[command] = true
		}
	}

	if len(commands) > count {
		commands = commands[len(commands)-count:]
	}

	return commands, nil
}

// isValidCommand checks if a command should be included
func isValidCommand(cmd string) bool {
	if len(cmd) <= 2 {
		return false
	}

	excludePatterns := []string{
		"go run main.go",
		"./cmd-makefile",
		"cmd-makefile",
		"./kmd",
		"kmd",
	}

	excludePrefixes := []string{
		"history",
		"fc -l",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(cmd, pattern) {
			return false
		}
	}

	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(cmd, prefix) {
			return false
		}
	}

	return true
}
