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
