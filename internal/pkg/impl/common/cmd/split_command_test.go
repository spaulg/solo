package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSplitCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SplitCommandTestSuite))
}

type SplitCommandTestSuite struct {
	suite.Suite
}

func (t *SplitCommandTestSuite) TestExecWithArg() {
	command := "/bin/ls -lh"

	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithDoubleQuotedArg() {
	command := "/bin/ls -lh \"/path with space\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path with space"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithSingleQuotedArg() {
	command := "/bin/ls -lh '/path with space'"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path with space"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEscapedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path with escaped spaces"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEscapedAndDoubleQuotedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEscapedAndSingleQuotedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces '/path with space'"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEmptyDoubleQuotedArg() {
	command := "/bin/ls -lh \"\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", ""}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEmptySingleQuotedArg() {
	command := "/bin/ls -lh ''"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", ""}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithDoubleEscapedArg() {
	command := "/bin/ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path\\", "with\\", "escaped\\", "spaces"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithNestedEscapedDoubleQuotedArg() {
	command := "/bin/ls -lh \"/path\\ with\\ space\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path\\ with\\ space"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithNestedEscapedSingleQuotedArg() {
	command := "/bin/ls -lh '/path\\ with\\ space'"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "/path\\ with\\ space"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEscapedDoubleQuoteArg() {
	command := "/bin/ls -lh \\\"\\\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "\"\""}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEscapedSingleQuoteArg() {
	command := "/bin/ls -lh \\'\\'"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "''"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithEscapedBackslashArg() {
	command := "/bin/ls -lh \\\\"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "\\"}, arguments)
}

func (t *SplitCommandTestSuite) TestExecWithQuotedEscapedSingleQuoteArg() {
	command := "/bin/ls -lh \"\\'\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/ls", command)
	t.Equal([]string{"-lh", "\\'"}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithoutArg() {
	command := "ls"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls"}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithArg() {
	command := "ls -lh"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh"}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithDoubleQuotedArg() {
	command := "ls -lh \"/path with space\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh \"/path with space\""}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithEscapedArg() {
	command := "ls -lh /path\\ with\\ escaped\\ spaces"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces"}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithEscapedAndDoubleQuotedArg() {
	command := "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithEmptyDoubleQuotedArg() {
	command := "ls -lh \"\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh \"\""}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithDoubleEscapedArg() {
	command := "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"}, arguments)
}

func (t *SplitCommandTestSuite) TestShellWithNestedEscapedDoubleQuotedArg() {
	command := "ls -lh \"/path\\ with\\ space\""
	command, arguments := SplitCommand(command)

	t.Equal("/bin/sh", command)
	t.Equal([]string{"-c", "ls -lh \"/path\\ with\\ space\""}, arguments)
}
