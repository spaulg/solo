package cmd

import (
	"strings"
)

type execCommandArgsExtractor struct {
	extracted    []string
	current      strings.Builder
	escaped      bool
	singleQuoted bool
	doubleQuoted bool
}

func newExecCommandArgsExtractor() *execCommandArgsExtractor {
	return &execCommandArgsExtractor{
		extracted:    []string{},
		current:      strings.Builder{},
		escaped:      false,
		singleQuoted: false,
		doubleQuoted: false,
	}
}

func (t *execCommandArgsExtractor) extractExecCommandArgs(command string) (string, []string) {
	for _, char := range command {
		t.extractExecCommandChar(char)
	}

	if t.current.Len() > 0 {
		t.extracted = append(t.extracted, t.current.String())
	}

	return t.extracted[0], t.extracted[1:]
}

func (t *execCommandArgsExtractor) extractExecCommandChar(char int32) {
	if t.isBackslash(char) {
		t.extractBackslash()
	} else if t.isDoubleQuote(char) {
		t.extractDoubleQuote()
	} else if t.isSingleQuote(char) {
		t.extractSingleQuote()
	} else if t.isSpace(char) {
		t.extractSpace()
	} else {
		t.extractChar(char)
	}
}

func (t *execCommandArgsExtractor) isBackslash(char int32) bool {
	return char == '\\' && !t.escaped && !t.singleQuoted && !t.doubleQuoted
}

func (t *execCommandArgsExtractor) extractBackslash() {
	t.escaped = true
}

func (t *execCommandArgsExtractor) isDoubleQuote(char int32) bool {
	return char == '"' && !t.escaped && !t.singleQuoted
}

func (t *execCommandArgsExtractor) extractDoubleQuote() {
	if t.doubleQuoted {
		t.doubleQuoted = false

		t.extracted = append(t.extracted, t.current.String())
		t.current.Reset()
	} else {
		t.doubleQuoted = true
	}
}

func (t *execCommandArgsExtractor) isSingleQuote(char int32) bool {
	return char == '\'' && !t.escaped && !t.doubleQuoted
}

func (t *execCommandArgsExtractor) extractSingleQuote() {
	if t.singleQuoted {
		t.singleQuoted = false

		t.extracted = append(t.extracted, t.current.String())
		t.current.Reset()
	} else {
		t.singleQuoted = true
	}
}

func (t *execCommandArgsExtractor) isSpace(char int32) bool {
	return char == ' ' && !t.escaped && !t.singleQuoted && !t.doubleQuoted
}

func (t *execCommandArgsExtractor) extractSpace() {
	if t.current.Len() > 0 {
		t.extracted = append(t.extracted, t.current.String())
		t.current.Reset()
	}
}

func (t *execCommandArgsExtractor) extractChar(char int32) {
	t.current.WriteRune(char)
	t.escaped = false
}
