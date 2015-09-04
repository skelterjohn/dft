package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

// This test file is intended to be read from top to bottom as a tutorial
// on dft filters.

// TestValue demonstrates simple inspections of an object, and prints it if
// the filter passes.
func TestValue(t *testing.T) {
	// test field values using .<field>=<value>
	testCase(t, tc{
		name:         "simple field cut",
		input:        `{"x":"y"}`,
		args:         []string{"f:.x=z"},
		expectedJSON: "",
	})
	testCase(t, tc{
		name:         "simple field match",
		input:        `{"x":"y"}`,
		args:         []string{"f:.x=y"},
		expectedJSON: `{"x":"y"}`,
	})

	// test index values using [<index>]=<value>
	testCase(t, tc{
		name:         "simple index cut",
		input:        `[1,2,3]`,
		args:         []string{"f:[1]=1"},
		expectedJSON: "",
	})
	testCase(t, tc{
		name:         "simple index match",
		input:        `[1,2,3]`,
		args:         []string{"f:[1]=2"},
		expectedJSON: `[1,2,3]`,
	})

	// test field or index presence by omitting the =
	testCase(t, tc{
		name:         "simple field presence match",
		input:        `{"x":"y"}`,
		args:         []string{"f:.x"},
		expectedJSON: `{"x":"y"}`,
	})
	testCase(t, tc{
		name:         "simple field presence cut",
		input:        `{"x":"y"}`,
		args:         []string{"f:.z"},
		expectedJSON: "",
	})
	testCase(t, tc{
		name:         "simple index presence match",
		input:        `[0,1,2]`,
		args:         []string{"f:[1]"},
		expectedJSON: `[0,1,2]`,
	})
	testCase(t, tc{
		name:         "simple index presence cut",
		input:        `[0,1,2]`,
		args:         []string{"f:[3]"},
		expectedJSON: "",
	})

	// nested fields and indices with [<index>].<field>, or .<field>[<index>]
	testCase(t, tc{
		name:         "nested cut, index then field",
		input:        `[{"x":"y"},{"x":"z"}]`,
		args:         []string{"f:[0].x=z"},
		expectedJSON: "",
	})
	testCase(t, tc{
		name:         "nested match, index then field",
		input:        `[{"x":"y"},{"x":"z"}]`,
		args:         []string{"f:[0].x=y"},
		expectedJSON: `[{"x":"y"},{"x":"z"}]`,
	})
	testCase(t, tc{
		name:         "nested match, field then index",
		input:        `{"x":[1,2,3],"y":["a", "b", "c"]}`,
		args:         []string{"f:.x[1]=2"},
		expectedJSON: `{"x":[1,2,3],"y":["a", "b", "c"]}`,
	})

	// what's matched can be a regular expression if you enclose with /<regexp>/
	testCase(t, tc{
		name:         "regexp field match",
		input:        `{"x": "abc123"}`,
		args:         []string{"f:.x=/.*c12.*$/"},
		expectedJSON: `{"x": "abc123"}`,
	})
	testCase(t, tc{
		name:         "regexp field cut",
		input:        `{"x": "abc123"}`,
		args:         []string{"f:.x=/.*c12$/"},
		expectedJSON: "",
	})

	// you can also compare to other values in the object
	testCase(t, tc{
		name:         "compare within object match",
		input:        `{"x":1,"y":1,"z":0}`,
		args:         []string{"f:.x=.y"},
		expectedJSON: `{"x":1,"y":1,"z":0}`,
	})
	testCase(t, tc{
		name:         "compare within object cut",
		input:        `{"x":1,"y":2,"z":0}`,
		args:         []string{"f:.x=.y"},
		expectedJSON: "",
	})

	// if you want to compare against a literal value that begins
	// with something that looks like an object lookup, you can
	// enclose it with quotes
	testCase(t, tc{
		name:         "quoted field match",
		input:        `{"x":".y","y":"something else"}`,
		args:         []string{`f:.x=".y"`},
		expectedJSON: `{"x":".y","y":"something else"}`,
	})
}

// TestExclusion demonstrates how to filter out part of an object.
func TestExclusion(t *testing.T) {
	// a list can be trimmed down using []
	testCase(t, tc{
		name:         "simple list exclusion",
		input:        `[1,2,3,2]`,
		args:         []string{"f:[]=2"},
		expectedJSON: `[2,2]`,
	})
	testCase(t, tc{
		name:         "nested list exclusion",
		input:        `[{"x":"y1", "w":"z1"},{"x":"y2", "w":"z2"}]`,
		args:         []string{"f:[].x=y1"},
		expectedJSON: `[{"w":"z1","x":"y1"}]`,
	})
	// even to nothing
	testCase(t, tc{
		name:         "complete list exclusion",
		input:        `[1,2,3,2]`,
		args:         []string{"f:[]=4"},
		expectedJSON: `[]`,
	})

	// fields can be filtered out as well using .()
	testCase(t, tc{
		name:         "simple field exclusion",
		input:        `{"x":1,"y":2,"z":2}`,
		args:         []string{"f:.()=2"},
		expectedJSON: `{"y":2,"z":2}`,
	})
	testCase(t, tc{
		name:         "complete field exclusion",
		input:        `{"x":1,"y":2,"z":2}`,
		args:         []string{"f:.()=3"},
		expectedJSON: `{}`,
	})
}

// TestExistence demonstrates how to pass an object through the filter if
// any part of it matches.
func TestExistence(t *testing.T) {
	// test if at least one index matches using [E]
	testCase(t, tc{
		name:         "simple index existence",
		input:        `[1,2,3,2]`,
		args:         []string{"f:[E]=2"},
		expectedJSON: `[1,2,3,2]`,
	})
	// test if at least one field matches using .(E)
	testCase(t, tc{
		name:         "simple field existence",
		input:        `{"x":1,"y":2,"z":2}`,
		args:         []string{"f:.(E)=2"},
		expectedJSON: `{"x":1,"y":2,"z":2}`,
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
		args:         []string{"f:[].x[E]=john"},
		expectedJSON: `[{"x":["john", "stephanie"], "y":"the best"}]`,
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
		expectedJSON: `
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

	// you can also use the { } operator to define a new root for
	// intra-object comparisons
	testCase(t, tc{
		name:         "compare within nested object",
		input:        `[{"x":1,"y":1,"z":0},{"x":1,"y":2,"z":1}]`,
		args:         []string{"f:[]{.x=.y}"},
		expectedJSON: `[{"x":1,"y":1,"z":0}]`,
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
		expectedJSON: `
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
		name:         "only the second element",
		input:        `[1,2,3,4,5]`,
		args:         []string{"f:@1"},
		expectedJSON: `[2]`,
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
		expectedJSON: `
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
	expectedJSON   string
	expectedError  string
	expectedOutput string
}

// dedent removes any common line indentation
func dedent(s string) string {
	var common []rune
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		leading := make([]rune, 100)[:0]
		for _, c := range line {
			if !unicode.IsSpace(c) {
				break
			}
			leading = append(leading, c)
		}
		if common == nil {
			common = leading
			continue
		}
		for i := range common {
			if common[i] != leading[i] {
				common = common[:i]
				break
			}
		}
	}
	var buf bytes.Buffer
	for _, line := range lines {
		if line == "" {
			fmt.Fprintln(&buf, line)
			continue
		}
		runes := []rune(line)[len(common):]
		fmt.Fprintf(&buf, "%s\n", string(runes))
	}
	return buf.String()
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

	got := strings.TrimSpace(buf.String())
	var want string

	if tc.expectedJSON != "" {
		want = strings.TrimSpace(regularizeJSON(tc.expectedJSON))
	}
	if tc.expectedOutput != "" {
		want = strings.TrimSpace(dedent(tc.expectedOutput))
	}

	if got != want {
		t.Errorf("for test %q: got\n%s\nwant:\n%s\n", tc.name, got, want)
		return false
	}
	return true
}
