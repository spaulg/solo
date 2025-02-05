package wms

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

const expectedStepName = "step name"
const expectedStepCommand = "/path/to/file"
const expectedWorkingDirectory = "/path/to/working/directory"

type StepTestSuite struct {
	suite.Suite
}

func TestStepTestSuite(t *testing.T) {
	suite.Run(t, new(StepTestSuite))
}

func (suite *StepTestSuite) TestStepAccessors() {
	step := NewStep(expectedStepName, expectedStepCommand, expectedWorkingDirectory)

	suite.Equal(expectedStepName, step.GetName())
	suite.Equal(expectedWorkingDirectory, *step.GetWorkingDirectory())
}

func (suite *StepTestSuite) TestExecWithoutArg() {
	step := NewStep(expectedStepName, expectedStepCommand, expectedWorkingDirectory)

	suite.Equal(expectedStepName, step.GetName())
	suite.Equal(expectedStepCommand, step.GetCommand())
	suite.Equal([]string{}, step.GetArguments())
	suite.Equal(expectedWorkingDirectory, *step.GetWorkingDirectory())
}

func (suite *StepTestSuite) TestExecWithArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithDoubleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \"/path with space\"", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithSingleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh '/path with space'", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\ with\\ escaped\\ spaces", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with escaped spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedAndDoubleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\"", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedAndSingleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\ with\\ escaped\\ spaces '/path with space'", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEmptyDoubleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \"\"", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", ""}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEmptySingleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh ''", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", ""}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithDoubleEscapedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh /path\\\\ with\\\\ escaped\\\\ spaces", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path\\", "with\\", "escaped\\", "spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithNestedEscapedDoubleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \"/path\\ with\\ space\"", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithNestedEscapedSingleQuotedArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh '/path\\ with\\ space'", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedDoubleQuoteArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \\\"\\\"", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "\"\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedSingleQuoteArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \\'\\'", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "''"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedBackslashArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \\\\", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "\\"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithQuotedEscapedSingleQuoteArg() {
	step := NewStep(expectedStepName, "/bin/ls -lh \"\\'\"", expectedWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "\\'"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithoutArg() {
	step := NewStep(expectedStepName, "ls", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithArg() {
	step := NewStep(expectedStepName, "ls -lh", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithDoubleQuotedArg() {
	step := NewStep(expectedStepName, "ls -lh \"/path with space\"", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh \"/path with space\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithEscapedArg() {
	step := NewStep(expectedStepName, "ls -lh /path\\ with\\ escaped\\ spaces", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithEscapedAndDoubleQuotedArg() {
	step := NewStep(expectedStepName, "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\"", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithEmptyDoubleQuotedArg() {
	step := NewStep(expectedStepName, "ls -lh \"\"", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh \"\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithDoubleEscapedArg() {
	step := NewStep(expectedStepName, "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithNestedEscapedDoubleQuotedArg() {
	step := NewStep(expectedStepName, "ls -lh \"/path\\ with\\ space\"", expectedWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh \"/path\\ with\\ space\""}, step.GetArguments())
}
