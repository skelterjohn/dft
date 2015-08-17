package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var errUnrecognizedOp = errors.New("unrecognized operation")

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
		if err == errUnrecognizedOp {
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
		return explicitIndex(obj, farg, index)
	}
	if field, ok := matchExactField(farg); ok {
		return explicitField(obj, farg, field)
	}

	if filters, ok := matchMulti(farg); ok {
		return filterMulti(obj, farg, filters)
	}

	return obj, errUnrecognizedOp
}

func matchExactIndex(farg string) (string, bool) {
	const (
		state_begin = iota
		state_open
		state_digit
		state_close
	)
	state := state_begin
	res := ""
loop:
	for _, c := range farg {
		switch state {
		case state_begin:
			if c == '[' {
				state = state_open
			} else {
				return "", false
			}
		case state_open:
			if unicode.IsDigit(c) {
				state = state_digit
				res += string(c)
			} else {
				return "", false
			}
		case state_digit:
			if unicode.IsDigit(c) {
				res += string(c)
				continue
			} else if c == ']' {
				state = state_close
				continue
			} else {
				return "", false
			}
		case state_close:
			break loop
		}
	}
	if state == state_close {
		_, err := strconv.Atoi(res)
		if err != nil {
			panic(fmt.Sprintf("somehow got bad atoi for %q: %v", res, err))
		}
		return res, true
	}
	return "", false
}

func matchExactField(farg string) (string, bool) {
	const (
		state_begin = iota
		state_open
		state_first
		state_rest
		state_close
	)
	state := state_begin
	res := ""
loop:
	for _, c := range farg {
		switch state {
		case state_begin:
			if c == '.' {
				state = state_open
			} else {
				return "", false
			}
		case state_open:
			if unicode.IsLetter(c) {
				state = state_first
				res += string(c)
			} else {
				return "", false
			}
		case state_first:
			if unicode.IsLetter(c) {
				state = state_rest
				res += string(c)
			} else {
				state = state_close
			}
		case state_rest:
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
				state = state_rest
				res += string(c)
			} else {
				state = state_close
			}
		case state_close:
			break loop
		}
	}
	return res, state == state_close
}

func matchMulti(farg string) ([]string, bool) {
	// log.Printf("mm: %q", farg)
	if !strings.HasPrefix(farg, "{") {
		return nil, false
	}
	depth := 1
	var filters []string
	var curfilter string
	for _, c := range farg[1:] {
		if depth == 0 {
			return nil, false
		}
		switch {
		case c == '{' && depth == 1:
			depth += 1
		case c == '}' && depth == 1:
			depth -= 1
		case c == ',' && depth == 1:
			filters = append(filters, curfilter)
		case c == '{' && depth > 1:
			curfilter += string(c)
			depth += 1
		case c == '}' && depth > 1:
			curfilter += string(c)
			depth -= 1
		case c == ',' && depth > 1:
			curfilter += string(c)
		default:
			curfilter += string(c)
		}
		if depth == 0 {
			filters = append(filters, curfilter)
		}
	}
	return filters, depth == 0
}

func filterExactValue(obj interface{}, farg string) (interface{}, error) {
	// log.Printf("fev: %v, %q", obj, farg)
	vstr := farg[1:]
	switch v := obj.(type) {
	case int:
		if fv, err := strconv.ParseInt(vstr, 10, 64); err == nil && int(fv) == v {
			return obj, nil
		}
	case float64:
		if fv, err := strconv.ParseFloat(vstr, 10); err == nil && float64(fv) == v {
			return obj, nil
		}
	case string:
		if vstr == v {
			return obj, nil
		}
	}
	return nil, errors.New("mismatch value")
}

func filterListExcludeMiss(obj interface{}, farg string) (interface{}, error) {
	// log.Printf("flem: %v, %q", obj, farg)
	rfarg := strings.TrimPrefix(farg, "[]")
	if v, ok := obj.([]interface{}); ok {
		var r []interface{}
		for _, subobj := range v {
			rsubobj, err := filter(subobj, rfarg)
			if err == nil {
				r = append(r, rsubobj)
			}
		}
		return r, nil
	}
	return nil, errors.New("not a list")
}

func filterListAtLeastOne(obj interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, "[E]")
	if v, ok := obj.([]interface{}); ok {
		for _, subobj := range v {
			if _, err := filter(subobj, rfarg); err == nil {
				return obj, nil
			}
		}
		return nil, fmt.Errorf("no matches for %q", rfarg)
	}
	return nil, errors.New("not a list")
}

func filterFieldsExcludeMiss(obj interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, ".()")
	if v, ok := obj.(map[string]interface{}); ok {
		r := map[string]interface{}{}
		for key, subobj := range v {
			rsubobj, err := filter(subobj, rfarg)
			if err == nil {
				r[key] = rsubobj
			}
		}
		return r, nil
	}
	return nil, errors.New("not a structure")
}

func filterFieldsAtLeastOne(obj interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, ".()")
	if v, ok := obj.(map[string]interface{}); ok {
		for _, subobj := range v {
			if _, err := filter(subobj, rfarg); err == nil {
				return obj, nil
			}
		}
		return nil, fmt.Errorf("no matches for %q", rfarg)
	}
	return nil, errors.New("not a structure")
}

func explicitIndex(obj interface{}, farg, index string) (interface{}, error) {
	// log.Printf("ei: %v, %q, %s", obj, farg, index)
	idx, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("somehow got bad strconv.ParseInt(%q): %v", index, err))
	}
	rfarg := strings.TrimPrefix(farg, fmt.Sprintf("[%s]", index))
	if v, ok := obj.([]interface{}); ok {
		subobj, err := filter(v[idx], rfarg)
		if err != nil {
			return nil, err
		}
		v[idx] = subobj
		return v, nil
	}
	return nil, errors.New("not a list")
}

func explicitField(obj interface{}, farg, field string) (interface{}, error) {
	// log.Printf("ef: %v, %q, %s", obj, farg, field)
	rfarg := strings.TrimPrefix(farg, fmt.Sprintf(".%s", field))
	if v, ok := obj.(map[string]interface{}); ok {
		subobj, err := filter(v[field], rfarg)
		if err != nil {
			return nil, err
		}
		v[field] = subobj
		return v, nil
	}
	return obj, nil
}

func filterMulti(obj interface{}, farg string, filters []string) (interface{}, error) {
	// log.Printf("fm: %q", filters)
	for _, f := range filters {
		var err error
		obj, err = filter(obj, f)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}
