package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexps = map[string]*regexp.Regexp{}
)

func filterExactValue(obj, root interface{}, farg string) (interface{}, error) {
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
		// quoted raw string comparison, to allow strings to begin with
		// one of the special prefixes: . [ /
		if strings.HasPrefix(vstr, `"`) && strings.HasSuffix(vstr, `"`) {
			vstr = vstr[1 : len(vstr)-1]
			if vstr == v {
				return obj, nil
			}
			return nil, errNotMatched
		}
		// regular expression
		if strings.HasPrefix(vstr, `/`) && strings.HasSuffix(vstr, `/`) {
			vstr = vstr[1 : len(vstr)-1]
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
		// default to raw string comparison
		if vstr == v {
			return obj, nil
		}
		return nil, errNotMatched

	}
	return nil, errNotMatched
}

func filterLookupValue(obj, root interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, "=")

	v, err := getValue(root, rfarg)
	if err != nil {
		return nil, err
	}

	if v == obj {
		return obj, nil
	}
	return nil, errNotMatched
}

func filterListExcludeMiss(obj, root interface{}, farg string) (interface{}, error) {
	// log.Printf("flem: %v, %q", obj, farg)
	rfarg := strings.TrimPrefix(farg, "[]")
	if v, ok := obj.([]interface{}); ok {
		r := make([]interface{}, 0, 0)
		for _, subobj := range v {
			rsubobj, err := filter(subobj, root, rfarg)
			if err == nil {
				r = append(r, rsubobj)
			}
		}
		return r, nil
	}
	return nil, errNotList
}

func filterListAtLeastOne(obj, root interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, "[E]")
	if v, ok := obj.([]interface{}); ok {
		for _, subobj := range v {
			if _, err := filter(subobj, root, rfarg); err == nil {
				return obj, nil
			}
		}
		return nil, errNotMatched
	}
	return nil, errNotList
}

func filterFieldsExcludeMiss(obj, root interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, ".()")
	if v, ok := obj.(map[string]interface{}); ok {
		r := map[string]interface{}{}
		for key, subobj := range v {
			rsubobj, err := filter(subobj, root, rfarg)
			if err == nil {
				r[key] = rsubobj
			}
		}
		return r, nil
	}
	return nil, errNotStruct
}

func filterFieldsAtLeastOne(obj, root interface{}, farg string) (interface{}, error) {
	rfarg := strings.TrimPrefix(farg, ".(E)")
	if v, ok := obj.(map[string]interface{}); ok {
		for _, subobj := range v {
			if _, err := filter(subobj, root, rfarg); err == nil {
				return obj, nil
			}
		}
		return nil, errNotMatched
	}
	return nil, errNotStruct
}

func filterExplicitIndex(obj, root interface{}, farg, index string) (interface{}, error) {
	// log.Printf("ei: %v, %q, %s", obj, farg, index)
	idx, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("somehow got bad strconv.ParseInt(%q): %v", index, err))
	}
	rfarg := strings.TrimPrefix(farg, fmt.Sprintf("[%s]", index))

	v, ok := obj.([]interface{})
	if !ok {
		return nil, errNotList
	}

	if rfarg == "" {
		if int(idx) < len(v) {
			return v, nil
		} else {
			return nil, errNotFound
		}
	}

	subobj, err := filter(v[idx], root, rfarg)
	if err != nil {
		return nil, err
	}
	v[idx] = subobj
	return v, nil
}

func filterExplicitField(obj, root interface{}, farg, field string) (interface{}, error) {
	// log.Printf("ef: %v, %q, %s", obj, farg, field)
	rfarg := strings.TrimPrefix(farg, fmt.Sprintf(".%s", field))

	v, ok := obj.(map[string]interface{})
	if !ok {
		return obj, nil
	}

	if rfarg == "" {
		if _, ok := v[field]; ok {
			return v, nil
		} else {
			return nil, errNotFound
		}
	}

	subobj, err := filter(v[field], root, rfarg)
	if err != nil {
		return nil, err
	}
	v[field] = subobj
	return v, nil
}

func filterMulti(obj, root interface{}, farg string, filters []string) (interface{}, error) {
	// log.Printf("fm: %q", filters)
	for _, f := range filters {
		var err error
		// the root of a multi's sub-expressions is this obj
		obj, err = filter(obj, obj, f)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func filterCut(obj, root interface{}, farg string, includes []string) (interface{}, error) {
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
