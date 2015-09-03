package main

import (
	"testing"
)

// This test file is intended to be read from top to bottom as a tutorial
// on dft output.

func TestTemplates(t *testing.T) {
	// inline a text/template with o:template=<format>
	testCase(t, tc{
		name:           "print every first element",
		input:          `[{"x":1},{"x":2}]`,
		args:           []string{`o:template={{range .}}{{printf "%v\n" .x}}{{end}}`},
		expectedOutput: "1\n2",
	})
	// you can use a file instead with o:templatefile=<file>, but it's
	// hard to have a unit test that explicitly uses the filesystem.
}
