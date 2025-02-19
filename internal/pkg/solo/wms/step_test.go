package wms

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

const expectedStepId = "abcde12345"
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
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, expectedStepCommand, &localWorkingDirectory)

	suite.Equal(expectedStepName, step.GetName())
	suite.Equal(expectedWorkingDirectory, step.GetWorkingDirectory())
}

func (suite *StepTestSuite) TestExecWithoutArg() {
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, expectedStepCommand, &localWorkingDirectory)

	suite.Equal(expectedStepName, step.GetName())
	suite.Equal(expectedStepCommand, step.GetCommand())
	suite.Equal([]string{}, step.GetArguments())
	suite.Equal(expectedWorkingDirectory, step.GetWorkingDirectory())
}

func (suite *StepTestSuite) TestExecWithArg() {
	command := "/bin/ls -lh"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithDoubleQuotedArg() {
	command := "/bin/ls -lh \"/path with space\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithSingleQuotedArg() {
	command := "/bin/ls -lh '/path with space'"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with escaped spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedAndDoubleQuotedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedAndSingleQuotedArg() {
	command := "/bin/ls -lh /path\\ with\\ escaped\\ spaces '/path with space'"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path with escaped spaces", "/path with space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEmptyDoubleQuotedArg() {
	command := "/bin/ls -lh \"\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", ""}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEmptySingleQuotedArg() {
	command := "/bin/ls -lh ''"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", ""}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithDoubleEscapedArg() {
	command := "/bin/ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path\\", "with\\", "escaped\\", "spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithNestedEscapedDoubleQuotedArg() {
	command := "/bin/ls -lh \"/path\\ with\\ space\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithNestedEscapedSingleQuotedArg() {
	command := "/bin/ls -lh '/path\\ with\\ space'"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "/path\\ with\\ space"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedDoubleQuoteArg() {
	command := "/bin/ls -lh \\\"\\\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "\"\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedSingleQuoteArg() {
	command := "/bin/ls -lh \\'\\'"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "''"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithEscapedBackslashArg() {
	command := "/bin/ls -lh \\\\"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "\\"}, step.GetArguments())
}

func (suite *StepTestSuite) TestExecWithQuotedEscapedSingleQuoteArg() {
	command := "/bin/ls -lh \"\\'\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/ls", step.GetCommand())
	suite.Equal([]string{"-lh", "\\'"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithoutArg() {
	command := "ls"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithArg() {
	command := "ls -lh"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithDoubleQuotedArg() {
	command := "ls -lh \"/path with space\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh \"/path with space\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithEscapedArg() {
	command := "ls -lh /path\\ with\\ escaped\\ spaces"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithEscapedAndDoubleQuotedArg() {
	command := "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh /path\\ with\\ escaped\\ spaces \"/path with space\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithEmptyDoubleQuotedArg() {
	command := "ls -lh \"\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh \"\""}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithDoubleEscapedArg() {
	command := "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh /path\\\\ with\\\\ escaped\\\\ spaces"}, step.GetArguments())
}

func (suite *StepTestSuite) TestShellWithNestedEscapedDoubleQuotedArg() {
	command := "ls -lh \"/path\\ with\\ space\""
	localWorkingDirectory := expectedWorkingDirectory
	step := NewStep(expectedStepId, expectedStepName, command, &localWorkingDirectory)

	suite.Equal("/bin/sh", step.GetCommand())
	suite.Equal([]string{"-c", "ls -lh \"/path\\ with\\ space\""}, step.GetArguments())
}
