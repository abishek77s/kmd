package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// Shell types
const (
	ShellBash        = "bash"
	ShellZsh         = "zsh"
	ShellFish        = "fish"
	ShellUnsupported = "unsupported"
)

// ShellConfig holds shell-specific configuration
type ShellConfig struct {
	Type        string
	HistoryFile string
	SetupLines  []string
}

// detectShell determines the current shell type and returns its configuration
func detectShell() (*ShellConfig, error) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return nil, fmt.Errorf("SHELL environment variable not set")
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	config := &ShellConfig{}

	switch {
	case strings.Contains(shellPath, "bash"):
		config.Type = ShellBash
		config.HistoryFile = getHistoryFilePath(usr.HomeDir, ".bash_history")
		config.SetupLines = []string{
			"# Auto-save history after each command (added by kmd)",
			"export PROMPT_COMMAND='history -a'",
			"export HISTSIZE=10000",
			"export HISTFILESIZE=20000",
			"shopt -s histappend",
		}

	case strings.Contains(shellPath, "zsh"):
		config.Type = ShellZsh
		config.HistoryFile = getHistoryFilePath(usr.HomeDir, ".zsh_history")
		config.SetupLines = []string{
			"# Auto-save history settings (added by kmd)",
			"setopt SHARE_HISTORY",
			"setopt HIST_SAVE_NO_DUPS",
			"setopt HIST_IGNORE_ALL_DUPS",
			"setopt HIST_FIND_NO_DUPS",
			"setopt HIST_IGNORE_SPACE",
			"setopt APPEND_HISTORY",
			"setopt INC_APPEND_HISTORY",
			"export HISTSIZE=10000",
			"export SAVEHIST=10000",
		}

	case strings.Contains(shellPath, "fish"):
		config.Type = ShellFish
		config.HistoryFile = getFishHistoryPath(usr.HomeDir)
		// Fish handles history automatically

	default:
		return nil, fmt.Errorf("unsupported shell: %s (supporting bash, zsh, fish)", shellPath)
	}

	return config, nil
}

// getHistoryFilePath checks HISTFILE env var first, then falls back to default
func getHistoryFilePath(homeDir, defaultFile string) string {
	if histFile, ok := os.LookupEnv("HISTFILE"); ok {
		if _, err := os.Stat(histFile); err == nil {
			return histFile
		}
	}
	return filepath.Join(homeDir, defaultFile)
}

// getFishHistoryPath returns the correct fish history file based on version
func getFishHistoryPath(homeDir string) string {
	oldHistFile := filepath.Join(homeDir, ".config", "fish", "fish_history")
	newHistFile := filepath.Join(homeDir, ".local", "share", "fish", "fish_history")

	// Try to get fish version
	cmd := exec.Command("fish", "--version")
	output, err := cmd.Output()
	if err != nil {
		return newHistFile // Default to newer location
	}

	version := strings.TrimSpace(string(output))
	if !strings.Contains(version, "version") {
		return newHistFile
	}

	// Parse version (fish, version 3.x.x)
	parts := strings.Fields(version)
	if len(parts) < 3 {
		return newHistFile
	}

	versionParts := strings.Split(parts[2], ".")
	if len(versionParts) < 2 {
		return newHistFile
	}

	major, errMajor := strconv.Atoi(versionParts[0])
	minor, errMinor := strconv.Atoi(versionParts[1])

	if errMajor != nil || errMinor != nil {
		return newHistFile
	}

	// Fish >= 2.3.0 uses new location
	if major > 2 || (major == 2 && minor >= 3) {
		return newHistFile
	}

	return oldHistFile
}

// setupHistoryConfig ensures history configuration is present in shell RC file
func setupHistoryConfig(shellConfig *ShellConfig) error {
	if shellConfig.Type == ShellFish {
		return nil // Fish doesn't need RC file setup
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	var rcFile string
	switch shellConfig.Type {
	case ShellBash:
		rcFile = filepath.Join(usr.HomeDir, ".bashrc")
	case ShellZsh:
		rcFile = filepath.Join(usr.HomeDir, ".zshrc")
	default:
		return nil
	}

	return ensureConfigInFile(rcFile, shellConfig.SetupLines)
}

// ensureConfigInFile adds configuration lines to RC file if not present
func ensureConfigInFile(rcFile string, lines []string) error {
	content, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	marker := "# kmd history config"
	if strings.Contains(string(content), marker) {
		return nil // Already configured
	}

	file, err := os.OpenFile(rcFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		file.WriteString("\n")
	}

	file.WriteString("\n" + marker + "\n")
	for _, line := range lines {
		file.WriteString(line + "\n")
	}
	file.WriteString("# end kmd config\n")

	return nil
}
