package main

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// This test file is intended to be read from top to bottom as a tutorial
// on dft filters.

// TestValue demonstrates simple inspections of an object, and prints it if
// the filter passes.
func TestValue(t *testing.T) {
	// test field values using .<field>=<value>
	testCase(t, tc{
		name:           "simple field cut",
		input:          `{"x":"y"}`,
		args:           []string{"f:.x=z"},
		expectedOutput: "",
	})
	testCase(t, tc{
		name:           "simple field match",
		input:          `{"x":"y"}`,
		args:           []string{"f:.x=y"},
		expectedOutput: `{"x":"y"}`,
	})

	// test index values using [<index>]=<value>
	testCase(t, tc{
		name:           "simple index cut",
		input:          `[1,2,3]`,
		args:           []string{"f:[1]=1"},
		expectedOutput: "",
	})
	testCase(t, tc{
		name:           "simple index match",
		input:          `[1,2,3]`,
		args:           []string{"f:[1]=2"},
		expectedOutput: `[1,2,3]`,
	})

	// nested fields and indices with [<index>].<field>, or .<field>[<index>]
	testCase(t, tc{
		name:           "nested cut, index then field",
		input:          `[{"x":"y"},{"x":"z"}]`,
		args:           []string{"f:[0].x=z"},
		expectedOutput: "",
	})
	testCase(t, tc{
		name:           "nested match, index then field",
		input:          `[{"x":"y"},{"x":"z"}]`,
		args:           []string{"f:[0].x=y"},
		expectedOutput: `[{"x":"y"},{"x":"z"}]`,
	})
	testCase(t, tc{
		name:           "nested match, field then index",
		input:          `{"x":[1,2,3],"y":["a", "b", "c"]}`,
		args:           []string{"f:.x[1]=2"},
		expectedOutput: `{"x":[1,2,3],"y":["a", "b", "c"]}`,
	})

	// what's matched can be a regular expression
	testCase(t, tc{
		name:           "regexp field match",
		input:          `{"x": "abc123"}`,
		args:           []string{"f:.x=.*c12.*$"},
		expectedOutput: `{"x": "abc123"}`,
	})
	testCase(t, tc{
		name:           "regexp field cut",
		input:          `{"x": "abc123"}`,
		args:           []string{"f:.x=.*c12$"},
		expectedOutput: "",
	})
}

// TestExclusion demonstrates how to filter out part of an object.
func TestExclusion(t *testing.T) {
	// a list can be trimmed down using []
	testCase(t, tc{
		name:           "simple list exclusion",
		input:          `[1,2,3,2]`,
		args:           []string{"f:[]=2"},
		expectedOutput: `[2,2]`,
	})
	testCase(t, tc{
		name:           "nested list exclusion",
		input:          `[{"x":"y1", "w":"z1"},{"x":"y2", "w":"z2"}]`,
		args:           []string{"f:[].x=y1"},
		expectedOutput: `[{"w":"z1","x":"y1"}]`,
	})
	// even to nothing
	testCase(t, tc{
		name:           "complete list exclusion",
		input:          `[1,2,3,2]`,
		args:           []string{"f:[]=4"},
		expectedOutput: `[]`,
	})

	// fields can be filtered out as well using .()
	testCase(t, tc{
		name:           "simple field exclusion",
		input:          `{"x":1,"y":2,"z":2}`,
		args:           []string{"f:.()=2"},
		expectedOutput: `{"y":2,"z":2}`,
	})
	testCase(t, tc{
		name:           "complete field exclusion",
		input:          `{"x":1,"y":2,"z":2}`,
		args:           []string{"f:.()=3"},
		expectedOutput: `{}`,
	})
}

// TestExistence demonstrates how to pass an object through the filter if
// any part of it matches.
func TestExistence(t *testing.T) {
	// test if at least one index matches using [E]
	testCase(t, tc{
		name:           "simple index existence",
		input:          `[1,2,3,2]`,
		args:           []string{"f:[E]=2"},
		expectedOutput: `[1,2,3,2]`,
	})
	// test if at least one field matches using .(E)
	testCase(t, tc{
		name:           "simple field existence",
		input:          `{"x":1,"y":2,"z":2}`,
		args:           []string{"f:.(E)=2"},
		expectedOutput: `{"x":1,"y":2,"z":2}`,
	})

	// nest an existence match under an exclusion match
	testCase(t, tc{
		name: "nested index existence",
		input: `
			[
			  {
			  	"x":["john", "stephanie"],
			  	"y":"the best"
			  },
			  {
			  	"x":["other","people"],
			  	"y":"the worst"
			  }
			]`,
		args:           []string{"f:[].x[E]=john"},
		expectedOutput: `[{"x":["john", "stephanie"], "y":"the best"}]`,
	})
}

// TestMulti lets you test multiple corresponding fields underneath other filters.
func TestMulti(t *testing.T) {
	// test multiple values by enclosing them with { and }
	// fun fact: an analog of this example is what made me create dft.
	testCase(t, tc{
		name: "key value check",
		input: `
			[
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"john"
			      },
			      {
			      	"key":"nonsense",
			        "value":"whatever"
			      }
			    ],
			    "useful":"information"
			  },
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"stephanie"
			      },
			      {
			      	"key":"nonsense",
			        "value":"do not care"
			      }
			    ],
			    "useful":"other information"
			  }
			]`,
		// read: in any item, there exists a meta item with the right key/val
		args: []string{"f:[].meta[E]{.key=name,.value=john}"},
		expectedOutput: `
			[
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"john"
			      },
			      {
			      	"key":"nonsense",
			        "value":"whatever"
			      }
			    ],
			    "useful":"information"
			  }
			]`,
	})
}

// TestCut demonstrates how to trim down based on index or field name, rather than value.
func TestCut(t *testing.T) {
	// allow only certain fields with @<field>
	testCase(t, tc{
		name: "only useful information",
		input: `
			[
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"john"
			      },
			      {
			      	"key":"nonsense",
			        "value":"whatever"
			      }
			    ],
			    "useful":"information"
			  },
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"stephanie"
			      },
			      {
			      	"key":"nonsense",
			        "value":"do not care"
			      }
			    ],
			    "useful":"other information"
			  }
			]`,
		args: []string{"f:[]@useful"},
		expectedOutput: `
			[
			  {
			    "useful":"information"
			  },
			  {
			    "useful":"other information"
			  }
			]`,
	})
	// same for indices
	testCase(t, tc{
		name:           "only the second element",
		input:          `[1,2,3,4,5]`,
		args:           []string{"f:@1"},
		expectedOutput: `[2]`,
	})
}

// TestManyFilters demonstrates sequential application of a filters.
func TestManyFilters(t *testing.T) {
	// fancy filter, but then we don't need the meta displayed in the output
	testCase(t, tc{
		name: "key value check",
		input: `
			[
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"john"
			      },
			      {
			      	"key":"nonsense",
			        "value":"whatever"
			      }
			    ],
			    "useful":"information"
			  },
			  {
			  	"meta":[
			      {
			      	"key":"name",
			        "value":"stephanie"
			      },
			      {
			      	"key":"nonsense",
			        "value":"do not care"
			      }
			    ],
			    "useful":"other information"
			  }
			]`,
		args: []string{
			"f:[].meta[E]{.key=name,.value=john}",
			"f:[]@useful",
		},
		expectedOutput: `
			[
			  {
			    "useful":"information"
			  }
			]`,
	})
}

// The tutorial ends here.

// What follows is code to make the tests easier to read

type tc struct {
	name           string
	input          string
	args           []string
	expectedOutput string
	expectedError  string
}

func regularizeJSON(s string) string {
	if s == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		panic(err)
	}
	return buf.String()
}

func testCase(t *testing.T, tc tc) bool {
	var buf bytes.Buffer
	if err := apply(strings.NewReader(tc.input), &buf, tc.args); err != nil {
		errRE := regexp.MustCompile(tc.expectedError)
		if errRE.FindStringIndex(err.Error()) == nil {
			t.Errorf("for test %q error: got %q, want %q", tc.name, err, tc.expectedError)
			return false
		}
	} else {
		if tc.expectedError != "" {
			t.Errorf("for test %q error: got nil, want %q", tc.name, tc.expectedError)
			return false
		}
	}
	eo := strings.TrimSpace(regularizeJSON(tc.expectedOutput))
	o := strings.TrimSpace(buf.String())
	if eo != o {
		t.Errorf("for test %q: got\n%s\nwant:\n%s\n", tc.name, o, eo)
		return false
	}
	return true
}
