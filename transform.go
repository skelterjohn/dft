package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func transformAllIndices(obj interface{}, targ string) (interface{}, error) {
	rtarg := strings.TrimPrefix(targ, "[]")
	if v, ok := obj.([]interface{}); ok {
		for i, sv := range v {
			if sr, err := transform(sv, rtarg); err == nil {
				v[i] = sr
			}
		}
		return v, nil
	}
	return nil, errors.New("not a list")
}

func transformExplicitIndex(obj interface{}, targ, index string) (interface{}, error) {
	rtarg := strings.TrimPrefix(targ, fmt.Sprintf("[%s]", index))
	idx, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("somehow got bad strconv.ParseInt(%q): %v", index, err))
	}
	if v, ok := obj.([]interface{}); ok {
		if sr, err := transform(v[idx], rtarg); err == nil {
			v[idx] = sr
		}
		return v, nil
	}
	return nil, errors.New("not a list")
}

func transformAllFields(obj interface{}, targ string) (interface{}, error) {
	rtarg := strings.TrimPrefix(targ, "()")
	if v, ok := obj.(map[string]interface{}); ok {
		for k, sv := range v {
			if sr, err := transform(sv, rtarg); err == nil {
				v[k] = sr
			}
		}
		return v, nil
	}
	return nil, errors.New("not a structure")
}

func transformExplicitField(obj interface{}, targ, field string) (interface{}, error) {
	rtarg := strings.TrimPrefix(targ, fmt.Sprintf(".%s", field))
	if v, ok := obj.(map[string]interface{}); ok {
		if sr, err := transform(v[field], rtarg); err == nil {
			v[field] = sr
		}
		return v, nil
	}
	return nil, errors.New("not a structure")
}
