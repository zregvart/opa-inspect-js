.PHONY: build
build:
	@GOOS=js GOARCH=wasm go build -o inspect.wasm
	@cp -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" .

demo: build
	@node example.js | jq .
