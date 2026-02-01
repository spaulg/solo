package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestStepTestSuite(t *testing.T) {
	suite.Run(t, new(StepTestSuite))
}

const expectedStepID = "abcde12345"
const expectedStepName = "step name"
const expectedStepCommand = "/path/to/file"
const expectedWorkingDirectory = "/path/to/working/directory"
const expectedShell = "/bin/sh"

type StepTestSuite struct {
	suite.Suite

	localWorkingDirectory string
}

func (t *StepTestSuite) SetupTest() {
	t.localWorkingDirectory = expectedWorkingDirectory
}

func (t *StepTestSuite) TestStepAccessors() {
	step := NewStep(expectedStepID, expectedStepName, expectedStepCommand, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedStepName, step.GetName())
	t.Equal(expectedWorkingDirectory, step.GetWorkingDirectory())
}

func (t *StepTestSuite) TestExecWithoutArg() {
	step := NewStep(expectedStepID, expectedStepName, expectedStepCommand, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedStepName, step.GetName())
	t.Equal(expectedStepCommand, step.GetCommand())
	t.Equal([]string{}, step.GetArguments())
	t.Equal(expectedWorkingDirectory, step.GetWorkingDirectory())
	t.Equal(expectedShell, step.GetShell())
}

func (t *StepTestSuite) TestExecWithArg() {
	command := "/bin/ls -lh"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithDoubleQuotedArg() {
	command := "/bin/ls -lh \"/path with space\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithSingleQuotedArg() {
	command := "/bin/ls -lh '/path with space'"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEscapedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path with escaped spaces"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEscapedAndDoubleQuotedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEscapedAndSingleQuotedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces '/path with space'"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEmptyDoubleQuotedArg() {
	command := "/bin/ls -lh \"\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", ""}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEmptySingleQuotedArg() {
	command := "/bin/ls -lh ''"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", ""}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithDoubleEscapedArg() {
	command := "/bin/ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path\\", "with\\", "escaped\\", "spaces"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithNestedEscapedDoubleQuotedArg() {
	command := "/bin/ls -lh \"/path\\ with\\ space\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithNestedEscapedSingleQuotedArg() {
	command := "/bin/ls -lh '/path\\ with\\ space'"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEscapedDoubleQuoteArg() {
	command := "/bin/ls -lh \\\"\\\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "\"\""}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEscapedSingleQuoteArg() {
	command := "/bin/ls -lh \\'\\'"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "''"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithEscapedBackslashArg() {
	command := "/bin/ls -lh \\\\"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "\\"}, step.GetArguments())
}

func (t *StepTestSuite) TestExecWithQuotedEscapedSingleQuoteArg() {
	command := "/bin/ls -lh \"\\'\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal("/bin/ls", step.GetCommand())
	t.Equal([]string{"-lh", "\\'"}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithoutArg() {
	command := "ls"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls"}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithArg() {
	command := "ls -lh"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh"}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithDoubleQuotedArg() {
	command := "ls -lh \"/path with space\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh \"/path with space\""}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithEscapedArg() {
	command := "ls -lh /path\\ with\\ escaped\\ spaces"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces"}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithEscapedAndDoubleQuotedArg() {
	command := "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithEmptyDoubleQuotedArg() {
	command := "ls -lh \"\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh \"\""}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithDoubleEscapedArg() {
	command := "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"}, step.GetArguments())
}

func (t *StepTestSuite) TestShellWithNestedEscapedDoubleQuotedArg() {
	command := "ls -lh \"/path\\ with\\ space\""
	step := NewStep(expectedStepID, expectedStepName, command, t.localWorkingDirectory, expectedShell)

	t.Equal(expectedShell, step.GetCommand())
	t.Equal(expectedShell, step.GetShell())
	t.Equal([]string{"-c", "ls -lh \"/path\\ with\\ space\""}, step.GetArguments())
}
