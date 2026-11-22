package subcommand

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// confirm prompts the user with a yes/no question until a valid answer is given.
// It returns nil on "y" and ErrUserAbortedCommand on "n" or when the input is
// closed (EOF), so a non-interactive or closed stdin cannot cause an endless loop.
func confirm(prompt string) error {
	for {
		var answer string

		fmt.Print(prompt + " (y/n)? ")
		_, err := fmt.Scanln(&answer)

		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			fmt.Println("n")
			return ErrUserAbortedCommand
		} else if err != nil {
			continue
		}

		switch strings.ToLower(answer) {
		case "y":
			return nil
		case "n":
			return ErrUserAbortedCommand
		}
	}
}
