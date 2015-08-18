package main

import (
	"fmt"
	"strconv"
	"strings"
)

func getExplicitIndex(obj interface{}, from, index string) (interface{}, error) {
	// log.Printf("gei %q %s", from, index)
	rfrom := strings.TrimPrefix(from, fmt.Sprintf("[%s]", index))
	idx, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("somehow got bad strconv.ParseInt(%q): %v", index, err))
	}
	if v, ok := obj.([]interface{}); ok {
		if int(idx) >= len(v) {
			return nil, errNotFound
		}
		if sr, err := getValue(v[idx], rfrom); err == nil {
			return sr, nil
		}
		return nil, errNotFound
	}
	return nil, errNotList

}

func getExplicitField(obj interface{}, from, field string) (interface{}, error) {
	rfrom := strings.TrimPrefix(from, fmt.Sprintf(".%s", field))
	if v, ok := obj.(map[string]interface{}); ok {
		if sv, ok := v[field]; ok {
			if sr, err := getValue(sv, rfrom); err == nil {
				return sr, nil
			}
			return nil, errNotFound
		}
		return nil, errNotFound
	}
	return nil, errNotStruct
}

func setExplicitIndex(obj interface{}, to string, setv interface{}, index string) (interface{}, error) {
	rto := strings.TrimPrefix(to, fmt.Sprintf("[%s]", index))
	idx, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("somehow got bad strconv.ParseInt(%q): %v", index, err))
	}

	v, ok := obj.([]interface{})
	if !ok {
		return nil, errNotList
	}

	// build up the list if it's not big enough
	for int(idx) >= len(v) {
		v = append(v, nil)
	}

	if rto == "" {
		v[idx] = setv
		return v, nil
	}

	if v[idx] == nil {
		if sv, err := newFieldObjRto(rto); err != nil {
			return nil, err
		} else {
			v[idx] = sv
		}
	}

	if sr, err := setValue(v[idx], rto, setv); err != nil {
		return nil, err
	} else {
		v[idx] = sr
		return v, nil
	}
}

func setExplicitField(obj interface{}, to string, setv interface{}, field string) (interface{}, error) {
	// log.Printf("sef %q %v %q", to, setv, field)
	rto := strings.TrimPrefix(to, fmt.Sprintf(".%s", field))
	v, ok := obj.(map[string]interface{})
	if !ok {
		return nil, errNotStruct
	}

	if rto == "" {
		v[field] = setv
		return v, nil
	}

	if _, ok := v[field]; !ok {
		if sv, err := newFieldObjRto(rto); err != nil {
			return nil, err
		} else {
			v[field] = sv
		}
	}

	if sr, err := setValue(v[field], rto, setv); err != nil {
		return nil, err
	} else {
		v[field] = sr
		return v, nil
	}
}

func newFieldObjRto(rto string) (interface{}, error) {
	if strings.HasPrefix(rto, ".") {
		return map[string]interface{}{}, nil
	}
	if strings.HasPrefix(rto, "[") {
		return make([]interface{}, 0, 1), nil
	}
	return nil, errIllegalOp
}
