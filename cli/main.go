package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var cmdCount int
	flag.IntVar(&cmdCount, "n", 5, "Number of commands to display from history")
	flag.IntVar(&cmdCount, "count", 5, "Number of commands to display from history")
	flag.Parse()

	// Validate input
	if cmdCount < 1 {
		cmdCount = 5
	}
	if cmdCount > 50 {
		cmdCount = 50
	}

	// Setup shell history configuration
	shellConfig, err := detectShell()
	if err != nil {
		fmt.Printf("%s %v\n", errorStyle.Render("Error:"), err)
		os.Exit(1)
	}

	if err := setupHistoryConfig(shellConfig); err != nil {
		fmt.Printf("%s Could not setup history configuration: %v\n", errorStyle.Render("Warning:"), err)
	}

	fmt.Println()

	// Run TUI
	p := tea.NewProgram(initialModel(cmdCount), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("%s %v\n", errorStyle.Render("Error:"), err)
		os.Exit(1)
	}
}
