package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	errUnrecognizedOp = errors.New("unrecognized operation")
	errIllegalOp      = errors.New("illegal operation")
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
}

func main() {
	var obj interface{}
	if err := json.NewDecoder(os.Stdin).Decode(&obj); err != nil {
		log.Fatalf("error reading stdin: %v", err)
	}

	for _, arg := range os.Args[1:] {
		var err error
		obj, err = ft(obj, arg)
		if err == errUnrecognizedOp || err == errIllegalOp {
			log.Fatalf("error with %q: %v", arg, err)
		}
		if err != nil {
			return
		}
	}

	if b, err := json.MarshalIndent(obj, "", "  "); err != nil {
		log.Fatalf("error marshalling: %v", err)
	} else {
		fmt.Printf("%s\n", b)
	}
}

func ft(obj interface{}, arg string) (interface{}, error) {
	if strings.HasPrefix(arg, "f:") {
		return filter(obj, strings.TrimPrefix(arg, "f:"))
	}
	if strings.HasPrefix(arg, "t:") {
		return transform(obj, strings.TrimPrefix(arg, "t:"))
	}
	return obj, errUnrecognizedOp
}

func filter(obj interface{}, farg string) (interface{}, error) {
	if farg == "" {
		return nil, errUnrecognizedOp
	}

	if strings.HasPrefix(farg, "=") {
		return filterExactValue(obj, farg)
	}

	if strings.HasPrefix(farg, "[]") {
		return filterListExcludeMiss(obj, farg)
	}
	if strings.HasPrefix(farg, "[E]") {
		return filterListAtLeastOne(obj, farg)
	}

	if strings.HasPrefix(farg, ".()") {
		return filterFieldsExcludeMiss(obj, farg)
	}
	if strings.HasPrefix(farg, ".(E)") {
		return filterFieldsAtLeastOne(obj, farg)
	}

	if index, ok := matchExactIndex(farg); ok {
		return filterExplicitIndex(obj, farg, index)
	}
	if field, ok := matchExactField(farg); ok {
		return filterExplicitField(obj, farg, field)
	}

	if filters, ok := matchMulti(farg); ok {
		return filterMulti(obj, farg, filters)
	}

	if includes, ok := matchCut(farg); ok {
		return filterCut(obj, farg, includes)
	}

	return obj, errUnrecognizedOp
}

func transform(obj interface{}, targ string) (interface{}, error) {
	// log.Printf("transform %q", targ)

	if targ == "" {
		return nil, errUnrecognizedOp
	}

	if strings.HasPrefix(targ, "[]") {
		return transformAllIndices(obj, targ)
	}
	if strings.HasPrefix(targ, ".()") {
		return transformAllFields(obj, targ)
	}

	if index, ok := matchExactIndex(targ); ok {
		return transformExplicitIndex(obj, targ, index)
	}
	if field, ok := matchExactField(targ); ok {
		return transformExplicitField(obj, targ, field)
	}

	if from, to, ok := matchReplace(targ); ok {
		return replace(obj, from, to)
	}

	return obj, nil
}

func replace(obj interface{}, from, to string) (interface{}, error) {
	// log.Printf("replace %q %q", from, to)
	v, err := getValue(obj, from)
	if err != nil {
		return nil, err
	}

	r, err := setValue(obj, to, v)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getValue(obj interface{}, from string) (interface{}, error) {
	// log.Printf("gv %q", from)
	if from == "" {
		return obj, nil
	}
	if index, ok := matchExactIndex(from); ok {
		return getExplicitIndex(obj, from, index)
	}
	if field, ok := matchExactField(from); ok {
		return getExplicitField(obj, from, field)
	}

	return nil, errIllegalOp
}

func setValue(obj interface{}, to string, v interface{}) (interface{}, error) {
	// log.Printf("sv %q %v", to, v)
	if to == "" {
		return v, nil
	}

	if index, ok := matchExactIndex(to); ok {
		return setExplicitIndex(obj, to, v, index)
	}
	if field, ok := matchExactField(to); ok {
		return setExplicitField(obj, to, v, field)
	}

	return nil, errIllegalOp
}

/*
$ dft \
  'f:[].metadata.items[E]{.key=foremanID,.value=foreman-not-on-borg-jasmuth}' \
  't:[]{.networkInterfaces[0].accessConfigs[0].natIP=.ip}' \
  'f:[]@name,ip' < in.txt
[
  {
    "name": "worker-ecba9d66-1c90-465c-8dd0-12e3ae867b66",
    "networkInterfaces": [
      {
        "accessConfigs": [
          {
            "natIP": "104.197.86.174"
          }
        ]
      }
    ]
  }
]

*/
