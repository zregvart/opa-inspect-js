build: .env index.js inspect.wasm wasm_exec.js

.PHONY: test
test: build node_modules
	@npm test

node_modules:
	@npm ci

.PHONY: clean
clean:
	@rm -rf node_modules inspect.wasm wasm_exec.js

.PHONY: demo
demo: build
	@node example.js |jq .

inspect.wasm: inspect.go
	@GOOS=js GOARCH=wasm go build -o inspect.wasm

wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	@cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" .

index.js: package.json node_modules
	@touch index.js

.env: .env.template
	@while read line; do eval "echo $${line}"; done < .env.template > .env
