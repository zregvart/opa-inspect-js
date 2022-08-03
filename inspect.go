package main

import (
	"bytes"
	"encoding/json"

	"github.com/gopherjs/gopherjs/js"
	"github.com/open-policy-agent/opa/ast"
)

func main() {
	js.Global.Set("Inspect", Inspect)
}

func Inspect(path, module string) string {
	mod, err := ast.ParseModuleWithOpts(path, module, ast.ParserOptions{ProcessAnnotation: true})
	if err != nil {
		return err.Error()
	}

	as, x := ast.BuildAnnotationSet([]*ast.Module{mod})
	if len(x) > 0 {
		return err.Error()
	}

	var buffy bytes.Buffer
	for _, rule := range mod.Rules {
		flattened := as.Chain(rule)
		for _, entry := range flattened {
			json.NewEncoder(&buffy).Encode(entry)
		}
	}

	return string(buffy.Bytes())
}
