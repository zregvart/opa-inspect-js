//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

	result := make([]*ast.AnnotationsRef, 0, len(mod.Rules))
	for _, rule := range mod.Rules {
		result = append(result, as.Chain(rule)...)
	}

	return result, nil
}

func tmp() ([]byte, error) {
	return make([]byte, 0, 10), nil
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

				var buffy bytes.Buffer
				if err := json.NewEncoder(&buffy).Encode(results); err != nil {
					ch <- result{err: err}
				}

				ch <- result{value: buffy.String()}
			}()

			return resolveWith(ch)
		} else if args[0].InstanceOf(js.Global().Get("Array")) {
			pathAry := args[0]
			len := pathAry.Length()

			ch := make(chan result)

			go func() {
				results := make([]*ast.AnnotationsRef, 0, len)

				for i := 0; i < len; i++ {
					path := pathAry.Index(i).String()

					moduleBytes, err := os.ReadFile(path)
					if err != nil {
						ch <- result{err: err}
					}

					module := string(moduleBytes)

					r, err := inspectSingle(path, module)
					if err != nil {
						ch <- result{err: err}
					}

					results = append(results, r...)
				}

				var buffy bytes.Buffer
				if err := json.NewEncoder(&buffy).Encode(results); err != nil {
					ch <- result{err: err}
				}

				ch <- result{value: buffy.String()}
			}()

			return resolveWith(ch)
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

			var buffy bytes.Buffer
			if err := json.NewEncoder(&buffy).Encode(results); err != nil {
				ch <- result{err: err}
			}

			ch <- result{value: buffy.String()}
		}()

		return resolveWith(ch)
	}

	return rejectWith(errors.New("at least one argument is required, given no arguments"))
}

func main() {
	o := js.Global().Get("Object").New()

	f := js.FuncOf(inspect)
	o.Set("inspect", f)

	o.Set("finish", js.FuncOf(func(this js.Value, args []js.Value) any {
		wait.Done()
		return nil
	}))

	p := unsafe.Pointer(&o)
	// fetching Value.ref which is at the top of the Value struct
	v := unsafe.Pointer(uintptr(p))
	that(*(*uint64)(v))

	wait.Add(1)

	wait.Wait()
}
