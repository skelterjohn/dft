package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexps = map[string]*regexp.Regexp{}
)

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

		if _, ok := regexps[vstr]; !ok {
			re, err := regexp.Compile(vstr)
			if err != nil {
				return nil, err
			}
			regexps[vstr] = re
		}
		re := regexps[vstr]
		if re.MatchString(v) {
			return obj, nil
		}
	}
	return nil, errors.New("mismatch value")
}

func filterListExcludeMiss(obj interface{}, farg string) (interface{}, error) {
	// log.Printf("flem: %v, %q", obj, farg)
	rfarg := strings.TrimPrefix(farg, "[]")
	if v, ok := obj.([]interface{}); ok {
		r := make([]interface{}, 0, 0)
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

func filterExplicitIndex(obj interface{}, farg, index string) (interface{}, error) {
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

func filterExplicitField(obj interface{}, farg, field string) (interface{}, error) {
	// log.Printf("ef: %v, %q, %s", obj, farg, field)
	rfarg := strings.TrimPrefix(farg, fmt.Sprintf(".%s", field))
	if v, ok := obj.(map[string]interface{}); ok {
		if rfarg == "" {
			if _, ok := v[field]; ok {
				return v, nil
			} else {
				return nil, errors.New("field not present")
			}
		}
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

func filterCut(obj interface{}, farg string, includes []string) (interface{}, error) {
	// log.Printf("fc: %v %q %q", obj, farg, includes)
	switch v := obj.(type) {
	case []interface{}:
		var r []interface{}
		for _, inc := range includes {
			index, err := strconv.ParseInt(inc, 10, 64)
			idx := int(index)
			if err != nil {
				return nil, errIllegalOp
			}
			if idx < len(v) {
				r = append(r, v[idx])
			}
		}
		return r, nil
	case map[string]interface{}:
		r := map[string]interface{}{}
		for _, inc := range includes {
			if sv, ok := v[inc]; ok {
				// log.Printf("including %q", inc)
				r[inc] = sv
			}
		}
		// log.Printf("fc ret %v", r)
		return r, nil
	default:
		return nil, errIllegalOp
	}
}
