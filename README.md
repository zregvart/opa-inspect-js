# `opa inspect` for JavaScript

This compiles the functionality of `opa inspect` wrapped in JavaScript by
`inspect.go` to WebAssembly, which is included in `index.js` with the Go runtime
from `wasm_exec.js` -- included in the Golang runtime and copied to the package.

## Example

Add the dependency:

```sh
npm add @zregvart/opa-inspect
```

Run this example with `node example.js`

```javascript
const opa = require("@zregvart/opa-inspect");

opa.inspect(
    "example.rego",
    `package example

# METADATA
# title: Task bundle was not used or is not defined
# description: |-
#   Check for existence of a task bundle. Enforcing this rule will
#   fail the contract if the task is not called from a bundle.
# custom:
#   short_name: disallowed_task_reference
#   failure_msg: Task '%s' does not contain a bundle reference
#
deny[msg] {
msg := "nope"
}`).then(json => {
  console.log(json);
});
```

## Building

Run `make build` to build, this copies `wasm_exec.js` from the Golang runtime
and compiles `inspect.go` to WebAssembly.

## Demo

Run `make demo` to build and run the example in `example.js`
