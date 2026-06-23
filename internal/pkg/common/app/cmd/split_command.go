package cmd

import "strings"

func SplitCommand(shell string, command string) (string, []string) {
	// An absolute command needs a leading slash
	// and at least one character for the binary
	if len(command) >= 2 && strings.HasPrefix(command, "/") {
		// Exec format
		return newExecCommandArgsExtractor().extractExecCommandArgs(command)
	}

	// Shell format
	return extractShellCommandArgs(shell, command)
}

func extractShellCommandArgs(shell string, command string) (string, []string) {
	return shell, []string{"-c", command}
}
