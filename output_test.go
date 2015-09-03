package main

import (
	"testing"
)

// This test file is intended to be read from top to bottom as a tutorial
// on dft output.

// TestTemplates demonstrates how to use a text/template for output
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

// TestManyObjects demonstrates how a series of json objects on input will
// result in a series of dft applications to output.
func TestManyObjects(t *testing.T) {

	// with many objects on input, the filter will be applied to each and
	// their results will be sent. for objects that don't match the filter,
	// there will be no output.
	testCase(t, tc{
		name: "repeated objects",
		input: `
			{"x": 1, "y":2, "z":1}
			{"x": 2, "y":2, "z":2}
			{"x": 3, "y":2, "z":3}
			{"x": 3, "y":3, "z":4}
			{"x": 1, "y":4, "z":5}
		`,
		args: []string{"f:.x=3"},
		expectedOutput: `
{
  "x": 3,
  "y": 2,
  "z": 3
}
{
  "x": 3,
  "y": 3,
  "z": 4
}
`,
	})

	// using templates you can make a list of json objects look like a nice table
	testCase(t, tc{
		name: "repeated formatted objects",
		input: `
			{"x": 1, "y":2, "z":1}
			{"x": 2, "y":2, "z":2}
			{"x": 3, "y":2, "z":3}
			{"x": 3, "y":3, "z":4}
			{"x": 1, "y":4, "z":5}
		`,
		args: []string{"f:.x=3", `o:template={{if .}}{{printf "%v %v\n" .y .z}}{{end}}`},
		expectedOutput: `
2 3
3 4
`,
	})
}
