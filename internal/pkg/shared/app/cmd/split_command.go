package cmd

import "strings"

func SplitCommand(shell string, command string) (string, []string) {
	if []rune(command)[0] == '/' {
		// Exec format
		return extractExecCommandArgs(command)
	}

	// Shell format
	return extractShellCommandArgs(command, shell)
}

func extractExecCommandArgs(command string) (string, []string) {
	var extracted []string
	var current strings.Builder
	escaped := false
	singleQuoted := false
	doubleQuoted := false

	for _, char := range command {
		if char == '\\' && !escaped && !singleQuoted && !doubleQuoted {
			escaped = true
		} else if char == '"' && !escaped && !singleQuoted {
			if doubleQuoted {
				doubleQuoted = false

				extracted = append(extracted, current.String())
				current.Reset()
			} else {
				doubleQuoted = true
			}
		} else if char == '\'' && !escaped && !doubleQuoted {
			if singleQuoted {
				singleQuoted = false

				extracted = append(extracted, current.String())
				current.Reset()
			} else {
				singleQuoted = true
			}
		} else if char == ' ' && !escaped && !singleQuoted && !doubleQuoted {
			if current.Len() > 0 {
				extracted = append(extracted, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
			escaped = false
		}
	}

	if current.Len() > 0 {
		extracted = append(extracted, current.String())
	}

	return extracted[0], extracted[1:]
}

func extractShellCommandArgs(command string, shell string) (string, []string) {
	return shell, []string{"-c", command}
}
