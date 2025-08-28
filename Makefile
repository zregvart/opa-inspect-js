build: .env main.js inspect.wasm wasm_exec.js

.PHONY: test-go
test-go:
	@PATH="$$PATH:$$(go env GOROOT)/lib/wasm" GOOS=js GOARCH=wasm go test ./... -timeout 1s

.PHONY: test-js
test-js: build node_modules
	@npm test

test: test-go test-js

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

wasm_exec.js: $(shell go env GOROOT)/lib/wasm/wasm_exec.js wasm_exec.js.patch
	@cp -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .
	@patch < wasm_exec.js.patch

main.js: package.json node_modules
	@touch main.js

.env: .env.template
	@while read line; do eval "echo $${line}"; done < .env.template > .env
