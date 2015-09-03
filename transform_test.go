package main

import (
	"testing"
)

// This test file is intended to be read from top to bottom as a tutorial
// on dft transforms.

// TestRename demonstrates simple top-level transforms.
func TestRename(t *testing.T) {
	// to copy a piece of the structure to elsewhere, use {<destination>=<source>}
	testCase(t, tc{
		name:           "copy x to y",
		input:          `{"x":"z","y":"w"}`,
		args:           []string{"t:{.y=.x}"},
		expectedOutput: `{"x":"z","y":"z"}`,
	})
	testCase(t, tc{
		name:           "copy 2 to 1",
		input:          `[1,2,3,4]`,
		args:           []string{"t:{[1]=[2]}"},
		expectedOutput: `[1,3,3,4]`,
	})

	// if the field isn't there, it will be added.
	testCase(t, tc{
		name:           "copy x to new y",
		input:          `{"x":"z"}`,
		args:           []string{"t:{.y=.x}"},
		expectedOutput: `{"x":"z","y":"z"}`,
	})
	// if the index isn't there, the list will be padded.
	// sorry, guessing that the new elements would be 0 in this case does not
	// apply generally (list elements don't have to be the same type), so we
	// get null instead.
	testCase(t, tc{
		name:           "copy 2 to 6",
		input:          `[1,2,3,4]`,
		args:           []string{"t:{[6]=[2]}"},
		expectedOutput: `[1,2,3,4,null,null,3]`,
	})
}

// TestDeepTransform demonstrates how to do transforms deep inside objects,
// potentially with the source value and destination value not being
// in the same place.
func TestDeepTransforms(t *testing.T) {
	// transforms first dig into the object, and the exact assignment goes
	// inside { and }.
	testCase(t, tc{
		name:           "nested transform",
		input:          `[{"x":2}]`,
		args:           []string{"t:[0]{.y=.x}"},
		expectedOutput: `[{"x":2,"y":2}]`,
	})
	// with unspecific indices or fields, using [] or .(), a transform
	// can be applied to many objects.
	testCase(t, tc{
		name:           "repeated index transform",
		input:          `[{"x":2}, {"x":4}]`,
		args:           []string{"t:[]{.y=.x}"},
		expectedOutput: `[{"x":2,"y":2},{"x":4,"y":4}]`,
	})
	testCase(t, tc{
		name:           "repeated field transform",
		input:          `{"x":[1,2,3], "y":[4,5,6]}`,
		args:           []string{"t:.(){[1]=[2]}"},
		expectedOutput: `{"x":[1,3,3], "y":[4,6,6]}`,
	})

	// you can dig in farther with the right hand side of the assignment.
	testCase(t, tc{
		name:           "transform with a deeper source",
		input:          `{"x":{"z": [1,2,3]}}`,
		args:           []string{"t:{.y=.x.z[1]}"},
		expectedOutput: `{"x":{"z": [1,2,3]},"y":2}`,
	})
	// or the left, too!
	testCase(t, tc{
		name:           "transform with a deeper destination",
		input:          `{"x":{"z": [1,2,3]}}`,
		args:           []string{"t:{.y.z=.x.z[1]}"},
		expectedOutput: `{"x":{"z": [1,2,3]},"y":{"z":2}}`,
	})
	// unless there is already something there that is incompatible.
	testCase(t, tc{
		name:          "want structure, got list",
		input:         `{"x":{"z": [1,2,3]},"y":[1,2,3]}`,
		args:          []string{"t:{.y.z=.x.z[1]}"},
		expectedError: "not a structure",
	})
	testCase(t, tc{
		name:          "got list, want structure",
		input:         `{"x":{"z": [1,2,3]},"y":"hello"}`,
		args:          []string{"t:{.y[1]=.x.z[1]}"},
		expectedError: "not a list",
	})

	// when you dig inside the { and }, it must be explicit.
	testCase(t, tc{
		name:          "no non-explicit indices",
		input:         `{"x":[1,2,3]}`,
		args:          []string{"t:{.y=.x[]}"},
		expectedError: `cannot use "\[\]" as source`,
	})
	testCase(t, tc{
		name:          "no non-explicit fields",
		input:         `{"x":[1,2,3]}`,
		args:          []string{"t:{.y=.()[1]}"},
		expectedError: `cannot use "\.\(\)\[1\]" as source`,
	})
}
