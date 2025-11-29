package main

import "github.com/charmbracelet/lipgloss"

// Styles using washed color palette
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8E8E8")).
			Background(lipgloss.Color("#4A5568")).
			Padding(0, 1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#68D391"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#63B3ED"))

	orderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B794F6"))

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9AE6B4"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FEB2B2"))

	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)
