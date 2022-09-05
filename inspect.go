package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/open-policy-agent/opa/ast"
)

var done = make(chan (bool))

func main() {
	js.Global().Set("opa", make(map[string]interface{}))

	opa := js.Global().Get("opa")
	f := js.FuncOf(func(this js.Value, args []js.Value) any {
		path := args[0].String()
		module := args[1].String()
		mod, err := ast.ParseModuleWithOpts(path, module, ast.ParserOptions{ProcessAnnotation: true})
		if err != nil {
			return "ERR: " + err.Error()
		}

		as, x := ast.BuildAnnotationSet([]*ast.Module{mod})
		if len(x) > 0 {
			return "ERR: " + err.Error()
		}

		var buffy bytes.Buffer
		for _, rule := range mod.Rules {
			flattened := as.Chain(rule)
			for _, entry := range flattened {
				if err := json.NewEncoder(&buffy).Encode(entry); err != nil {
					return "ERR: " + err.Error()
				}
			}
		}

		return string(buffy.Bytes())
	})
	opa.Set("inspect", f)

	opa.Set("finish", js.FuncOf(func(this js.Value, args []js.Value) any {
		done <- true
		return nil
	}))
	<-done
}
