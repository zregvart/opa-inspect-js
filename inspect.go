//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall/js"
	"unsafe"

	"github.com/open-policy-agent/opa/ast"
)

var wait sync.WaitGroup

func that(uint64)

type result struct {
	value string
	err   error
}

func rejectWith(err error) js.Value {
	return js.Global().Get("Promise").Call("reject", err.Error())
}

func resolveWith(r chan result) js.Value {
	wait.Add(1)
	return js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		reject := args[1]

		go func() {
			defer wait.Done()
			res := <-r
			if res.err != nil {
				reject.Invoke(res.err.Error())
			} else {
				resolve.Invoke(res.value)
			}
		}()

		return nil
	}))
}

func inspectSingle(path, module string) ([]*ast.AnnotationsRef, error) {
	mod, err := ast.ParseModuleWithOpts(path, module, ast.ParserOptions{ProcessAnnotation: true})
	if err != nil {
		return nil, err
	}

	as, x := ast.BuildAnnotationSet([]*ast.Module{mod})
	if len(x) > 0 {
		return nil, err
	}

	results := make([]*ast.AnnotationsRef, 0, len(mod.Rules))
	for _, rule := range mod.Rules {
		results = append(results, as.Chain(rule)...)
	}

	return results, nil
}

func inspectMultiple(paths, modules []string) ([]*ast.AnnotationsRef, error) {
	noPaths := len(paths)
	if noPaths != len(modules) {
		return nil, fmt.Errorf("given uneven number of paths and modules: %d != %d", noPaths, len(modules))
	}

	results := make([]*ast.AnnotationsRef, 0, noPaths)
	for i := 0; i < noPaths; i++ {
		r, err := inspectSingle(paths[i], modules[i])
		if err != nil {
			return nil, err
		}

		results = append(results, r...)
	}

	return results, nil
}

func serialize(results []*ast.AnnotationsRef) (string, error) {
	var buffy bytes.Buffer
	if err := json.NewEncoder(&buffy).Encode(results); err != nil {
		return "", err
	}

	return buffy.String(), nil
}

func inspect(this js.Value, args []js.Value) any {
	if len(args) == 1 {
		if args[0].Type() == js.TypeString {
			// given a single path
			path := args[0].String()

			ch := make(chan result)

			go func() {
				moduleBytes, err := os.ReadFile(path)
				if err != nil {
					ch <- result{err: err}
				}

				module := string(moduleBytes)

				results, err := inspectSingle(path, module)
				if err != nil {
					ch <- result{err: err}
				}

				if json, err := serialize(results); err != nil {
					ch <- result{err: err}
				} else {
					ch <- result{value: json}
				}
			}()

			return resolveWith(ch)
		} else if args[0].InstanceOf(js.Global().Get("Array")) {
			pathAry := args[0]
			len := pathAry.Length()

			ch := make(chan result)

			go func() {
				paths := make([]string, 0, len)
				modules := make([]string, 0, len)

				for i := 0; i < len; i++ {
					path := pathAry.Index(i).String()
					paths = append(paths, path)

					moduleBytes, err := os.ReadFile(path)
					if err != nil {
						ch <- result{err: err}
					}

					module := string(moduleBytes)
					modules = append(modules, module)
				}

				if results, err := inspectMultiple(paths, modules); err != nil {
					ch <- result{err: err}
				} else if json, err := serialize(results); err != nil {
					ch <- result{err: err}
				} else {
					ch <- result{value: json}
				}
			}()

			return resolveWith(ch)
		} else if pipe := args[0].Get("pipe"); pipe.Type() == js.TypeFunction {
			nop := js.FuncOf(func(this js.Value, args []js.Value) any {
				return nil
			})

			ch := make(chan result)

			go func() {
				paths := make([]string, 0)
				modules := make([]string, 0)

				stream := js.ValueOf(map[string]any{
					"on":   nop,
					"once": nop,
					"emit": nop,
					"write": js.FuncOf(func(this js.Value, args []js.Value) any {
						paths = append(paths, args[0].Get("path").String())
						buff := args[0].Get("contents")
						module := buff.Call("toString").String()
						modules = append(modules, module)

						return nil
					}),
					"end": js.FuncOf(func(this js.Value, args []js.Value) any {
						if results, err := inspectMultiple(paths, modules); err != nil {
							ch <- result{err: err}
						} else if json, err := serialize(results); err != nil {
							ch <- result{err: err}
						} else {
							ch <- result{value: json}
						}

						return nil
					}),
				})
				args[0].Call("pipe", stream)
			}()

			return resolveWith(ch)
		} else {
			return rejectWith(fmt.Errorf("unsupported argument: %s", js.Global().Get("Object").Call("getPrototypeOf", args[0])))
		}
	}

	if len(args) == 2 {
		if args[0].Type() != js.TypeString || args[1].Type() != js.TypeString {
			return rejectWith(errors.New("when given two arguments expecting both to be of string type"))
		}

		// given a path and module (Rego source) as string
		path := args[0].String()
		module := args[1].String()

		ch := make(chan result)

		go func() {
			results, err := inspectSingle(path, module)
			if err != nil {
				ch <- result{err: err}
			}

			if json, err := serialize(results); err != nil {
				ch <- result{err: err}
			} else {
				ch <- result{value: json}
			}
		}()

		return resolveWith(ch)
	}

	return rejectWith(errors.New("at least one argument is required, given no arguments"))
}

func main() {
	o := js.ValueOf(map[string]any{
		"inspect": js.FuncOf(inspect),
		"finish": js.FuncOf(func(this js.Value, args []js.Value) any {
			wait.Done()
			return nil
		}),
	})

	p := unsafe.Pointer(&o)
	// fetching Value.ref which is at the top of the Value struct
	v := unsafe.Pointer(uintptr(p))
	that(*(*uint64)(v))

	wait.Add(1)

	wait.Wait()
}
