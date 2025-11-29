package main

import (
	"fmt"
	"os"
	"time"
)

type makefileMsg struct {
	success bool
	error   error
}

func generateMakefile(choices []string, order []int) makefileMsg {
	if len(order) == 0 {
		return makefileMsg{false, fmt.Errorf("no commands selected")}
	}

	content := "# Generated Makefile from command history\n"
	content += "# Created with kmd tool\n"
	content += fmt.Sprintf("# Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	content += ".PHONY: all clean"

	for i := range order {
		content += fmt.Sprintf(" step%d", i+1)
	}
	content += "\n\n"
	content += "all:"
	for i := range order {
		content += fmt.Sprintf(" step%d", i+1)
	}
	content += "\n\n"

	for i, cmdIdx := range order {
		stepName := fmt.Sprintf("step%d", i+1)
		command := choices[cmdIdx]

		content += fmt.Sprintf("%s:\n", stepName)
		content += fmt.Sprintf("\t@echo \"🔧 Step %d: %s\"\n", i+1, command)
		content += fmt.Sprintf("\t%s\n\n", command)
	}

	content += "clean:\n"
	content += "\t@echo \"🧹 Cleaning up...\"\n"
	content += "\t# Add cleanup commands here\n"

	err := os.WriteFile("Makefile", []byte(content), 0644)
	return makefileMsg{success: err == nil, error: err}
}
