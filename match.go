package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

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
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
				state = state_first
				res += string(c)
			} else {
				return "", false
			}
		case state_first:
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
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
	return res, state != state_open
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
			curfilter = ""
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

func matchCut(farg string) ([]string, bool) {
	if !strings.HasPrefix(farg, "@") {
		return nil, false
	}
	return strings.Split(farg[1:], ","), true
}

func matchReplace(targ string) (to, from string, ok bool) {
	if !strings.HasPrefix(targ, "{") {
		return "", "", false
	}
	if !strings.HasSuffix(targ, "}") {
		return "", "", false
	}
	tokens := strings.Split(targ[1:len(targ)-1], "=")
	if len(tokens) != 2 {
		return "", "", false
	}
	return tokens[0], tokens[1], true
}
