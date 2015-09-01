package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
)

var (
	errUnrecognizedOp = errors.New("unrecognized operation")
	errIllegalOp      = errors.New("illegal operation")

	errNotList    = errors.New("not a list")
	errNotStruct  = errors.New("not a structure")
	errNotFound   = errors.New("not found")
	errNotMatched = errors.New("values do not match")
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
}

func apply(in io.Reader, out io.Writer, args []string) error {
	dec := json.NewDecoder(in)

	for {
		var obj interface{}
		if err := dec.Decode(&obj); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading stdin: %v", err)
		}

		for i, arg := range args {
			var err error
			obj, err = ft(obj, arg)
			// if err == errUnrecognizedOp || err == errIllegalOp {
			if err != nil {
				log.Fatalf("error with %q: %v", arg, err)
				continue
			}
			// }
			// if err != nil {
			// 	continue
			// }

			if obj == nil {
				if i != len(args)-1 {
					log.Printf("unused args: %q", args[i+1:])
				}
				break
			}
		}

		if obj == nil {
			continue
		}

		if b, err := json.MarshalIndent(obj, "", "  "); err != nil {
			return fmt.Errorf("error marshalling: %v", err)
		} else {
			fmt.Fprintf(out, "%s\n", b)
		}
	}
	return nil
}

func main() {
	if err := apply(os.Stdin, os.Stdout, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func ft(obj interface{}, arg string) (interface{}, error) {
	switch {
	case strings.HasPrefix(arg, "f:"):
		return filter(obj, strings.TrimPrefix(arg, "f:"))
	case strings.HasPrefix(arg, "t:"):
		return transform(obj, strings.TrimPrefix(arg, "t:"))
	case strings.HasPrefix(arg, "o:"):
		return nil, output(obj, strings.TrimPrefix(arg, "o:"))
	default:
		return nil, errUnrecognizedOp
	}
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

	if to, from, ok := matchReplace(targ); ok {
		return replace(obj, to, from)
	}

	return nil, errUnrecognizedOp
}

func output(obj interface{}, oarg string) error {
	switch {
	case strings.HasPrefix(oarg, "templatefile="):
		return printTemplateFile(obj, strings.TrimPrefix(oarg, "templatefile="))
	case strings.HasPrefix(oarg, "template="):
		return printTemplate(obj, strings.TrimPrefix(oarg, "template="))
	default:
		return errUnrecognizedOp
	}
}

func replace(obj interface{}, to, from string) (interface{}, error) {
	// log.Printf("replace %q %q", from, to)
	v, err := getValue(obj, from)
	if err != nil {
		return nil, err
	}

	r, err := setValue(obj, to, v)
	if err != nil {
		return nil, fmt.Errorf("could not set %q: %v", to, err)
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

func printTemplateFile(obj interface{}, filename string) error {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, obj)
}

func printTemplate(obj interface{}, format string) error {
	tmpl, err := template.New("dft").Parse(format)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, obj)
}
