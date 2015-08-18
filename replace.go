package main

import (
	"errors"
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
			return nil, errors.New("index not found")
		}
		if sr, err := getValue(v[idx], rfrom); err == nil {
			return sr, nil
		}
		return nil, errors.New("value not found")
	}
	return nil, errors.New("not a list")

}

func getExplicitField(obj interface{}, from, field string) (interface{}, error) {
	rfrom := strings.TrimPrefix(from, fmt.Sprintf(".%s", field))
	if v, ok := obj.(map[string]interface{}); ok {
		if sv, ok := v[field]; ok {
			if sr, err := getValue(sv, rfrom); err == nil {
				return sr, nil
			}
			return nil, errors.New("value not found")
		}
		return nil, errors.New("field not found")
	}
	return nil, errors.New("not a structure")
}

func setExplicitIndex(obj interface{}, to string, setv interface{}, index string) (interface{}, error) {
	return obj, nil
}

func setExplicitField(obj interface{}, to string, setv interface{}, field string) (interface{}, error) {
	// log.Printf("sef %q %v %q", to, setv, field)
	rto := strings.TrimPrefix(to, fmt.Sprintf(".%s", field))
	v, ok := obj.(map[string]interface{})
	if !ok {
		return nil, errors.New("not a structure")
	}
	var sv interface{}
	var fok bool
	if sv, fok = v[field]; !fok {
		if rto == "" {
			v[field] = setv
			return v, nil
		}
		var err error
		sv, err = newFieldObjRto(rto)
		if err != nil {
			return nil, err
		}
		v[field] = sv
	}
	if sr, err := setValue(sv, rto, setv); err == nil {
		v[field] = sr
		return v, nil
	}
	return nil, errors.New("value not found")
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
