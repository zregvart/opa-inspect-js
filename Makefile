.PHONY: build
build:
	@GOPHERJS_GOROOT="$$(go env GOROOT)" go run github.com/gopherjs/gopherjs build --tags safe .

demo: build
	@node example.js | jq .
