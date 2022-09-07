package main

import (
	"bytes"
	"encoding/json"
	"os"
	"syscall/js"

	"github.com/open-policy-agent/opa/ast"
)

var done = make(chan (bool))

type readFn func(path string) ([]byte, error)

func inspectSingle(path, module string) ([]*ast.AnnotationsRef, error) {
	mod, err := ast.ParseModuleWithOpts(path, module, ast.ParserOptions{ProcessAnnotation: true})
	if err != nil {
		return nil, err
	}

	as, x := ast.BuildAnnotationSet([]*ast.Module{mod})
	if len(x) > 0 {
		return nil, err
	}

	result := make([]*ast.AnnotationsRef, 0, len(mod.Rules))
	for _, rule := range mod.Rules {
		result = append(result, as.Chain(rule)...)
	}

	return result, nil
}

func determineReadFunc(this js.Value) readFn {
	r := this.Get("read")
	if r.Type() == js.TypeFunction {
		return func(path string) ([]byte, error) {
			val := r.Invoke(path)
			bytes := make([]byte, val.Length())

			if js.CopyBytesToGo(bytes, val) == 0 {
				panic("no bytes copied")
			}

			return bytes, nil
		}
	}

	return os.ReadFile
}

func inspect(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return "ERR: path argument is required, given no arguments"
	}

	var paths []string
	if args[0].Type() == js.TypeString {
		paths = []string{args[0].String()}
	} else if len := args[0].Length(); len > 0 {
		for i := 0; i < len; i++ {
			paths = append(paths, args[0].Index(i).String())
		}
	}

	read := determineReadFunc(this)

	var modules []string
	if len(args) == 2 && args[1].Type() == js.TypeString {
		if len(paths) > 1 {
			return "ERR: provided more than one filename and a single Rego module"
		}
		modules = []string{args[1].String()}
	} else {
		for _, path := range paths {
			if bytes, err := read(path); err == nil {
				modules = append(modules, string(bytes))
			} else {
				return "ERR: " + err.Error()
			}
		}
	}

	var results []*ast.AnnotationsRef

	for i := 0; i < len(paths); i++ {
		path := paths[i]
		module := modules[i]
		result, err := inspectSingle(path, module)
		if err != nil {
			return "ERR: " + err.Error()
		}

		results = append(results, result...)
	}

	var buffy bytes.Buffer
	if err := json.NewEncoder(&buffy).Encode(results); err != nil {
		return "ERR: " + err.Error()
	}

	return string(buffy.Bytes())
}

func main() {
	js.Global().Set("opa", make(map[string]interface{}))

	opa := js.Global().Get("opa")
	f := js.FuncOf(inspect)
	opa.Set("inspect", f)

	opa.Set("finish", js.FuncOf(func(this js.Value, args []js.Value) any {
		done <- true
		return nil
	}))
	<-done
}
