package wms

import "testing"
import asserter "github.com/stretchr/testify/assert"

const expectedStepName = "step name"
const expectedStepCommand = "/path/to/file"
const expectedWorkingDirectory = "/path/to/working/directory"

func TestStepAccessors(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, expectedStepCommand, expectedWorkingDirectory)

	assert.Equal(expectedStepName, step.GetName())
	assert.Equal(expectedWorkingDirectory, *step.GetWorkingDirectory())
}

func TestExecWithoutArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, expectedStepCommand, expectedWorkingDirectory)

	assert.Equal(expectedStepName, step.GetName())
	assert.Equal(expectedStepCommand, step.GetCommand())
	assert.Equal([]string{}, step.GetArguments())
	assert.Equal(expectedWorkingDirectory, *step.GetWorkingDirectory())
}

func TestExecWithArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh"}, step.GetArguments())
}

func TestExecWithDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \"/path with space\"", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func TestExecWithSingleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh '/path with space'", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func TestExecWithEscapedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\ with\\ escaped\\ spaces", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path with escaped spaces"}, step.GetArguments())
}

func TestExecWithEscapedAndDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\"", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func TestExecWithEscapedAndSingleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\ with\\ escaped\\ spaces '/path with space'", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func TestExecWithEmptyDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \"\"", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", ""}, step.GetArguments())
}

func TestExecWithEmptySingleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh ''", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", ""}, step.GetArguments())
}

func TestExecWithDoubleEscapedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\\\ with\\\\ escaped\\\\ spaces", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path\\", "with\\", "escaped\\", "spaces"}, step.GetArguments())
}

func TestExecWithNestedEscapedDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \"/path\\ with\\ space\"", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func TestExecWithNestedEscapedSingleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh '/path\\ with\\ space'", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func TestExecWithEscapedDoubleQuoteArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \\\"\\\"", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "\"\""}, step.GetArguments())
}

func TestExecWithEscapedSingleQuoteArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \\'\\'", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "''"}, step.GetArguments())
}

func TestExecWithEscapedBackslashArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \\\\", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "\\"}, step.GetArguments())
}

func TestExecWithQuotedEscapedSingleQuoteArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "/bin/ls -lh \"\\'\"", expectedWorkingDirectory)

	assert.Equal("/bin/ls", step.GetCommand())
	assert.Equal([]string{"-lh", "\\'"}, step.GetArguments())
}

func TestShellWithoutArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls"}, step.GetArguments())
}

func TestShellWithArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh"}, step.GetArguments())
}

func TestShellWithDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh \"/path with space\"", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh \"/path with space\""}, step.GetArguments())
}

func TestShellWithEscapedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh /path\\ with\\ escaped\\ spaces", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces"}, step.GetArguments())
}

func TestShellWithEscapedAndDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\"", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""}, step.GetArguments())
}

func TestShellWithEmptyDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh \"\"", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh \"\""}, step.GetArguments())
}

func TestShellWithDoubleEscapedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"}, step.GetArguments())
}

func TestShellWithNestedEscapedDoubleQuotedArg(t *testing.T) {
	assert := asserter.New(t)
	step := NewStep(expectedStepName, "ls -lh \"/path\\ with\\ space\"", expectedWorkingDirectory)

	assert.Equal("/bin/sh", step.GetCommand())
	assert.Equal([]string{"-c", "ls -lh \"/path\\ with\\ space\""}, step.GetArguments())
}
