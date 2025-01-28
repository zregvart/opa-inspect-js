package main

import (
	"errors"
	"syscall/js"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

var stubThat = js.ValueOf(map[string]any{})

var rego = `# METADATA
# title: title
# description: description
package example

# METADATA
# title: Task bundle was not used or is not defined
# description: |-
#   Check for existence of a task bundle. Enforcing this rule will
#   fail the contract if the task is not called from a bundle.
# custom:
#   short_name: disallowed_task_reference
#   failure_msg: Task '%s' does not contain a bundle reference
#
deny contains msg if {
	msg := "nope"
}
`

func outcome(p any) (string, error) {
	promise := p.(js.Value)

	ch := make(chan result)
	go func() {
		promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			ch <- result{value: args[0].String()}

			return nil
		})).
			Call("catch", js.FuncOf(func(this js.Value, args []js.Value) any {
				ch <- result{err: errors.New(args[0].String())}

				return nil
			}))
	}()

	r := <-ch

	return r.value, r.err
}

func TestInspectProvidedFile(t *testing.T) {
	args := []js.Value{
		js.ValueOf("example.rego"),
		js.ValueOf(rego),
	}

	json, err := outcome(inspect(stubThat, args))
	assert.NoError(t, err)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoaded(t *testing.T) {
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
	}

	json, err := outcome(inspect(stubThat, args))
	assert.NoError(t, err)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedSecondArgumentNull(t *testing.T) {
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
		js.Null(),
	}

	json, err := outcome(inspect(stubThat, args))
	assert.EqualError(t, err, "when given two arguments expecting both to be of string type")

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedSecondArgumentUndefined(t *testing.T) {
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
		js.Undefined(),
	}

	json, err := outcome(inspect(stubThat, args))
	assert.EqualError(t, err, "when given two arguments expecting both to be of string type")

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileGivenAsArray(t *testing.T) {
	ary := js.Global().Get("Array").New("__test__/example.rego")
	args := []js.Value{
		ary,
	}

	json, err := outcome(inspect(stubThat, args))
	assert.NoError(t, err)

	cupaloy.SnapshotT(t, json)
}

func TestInspectMultipleFilesGivenAsArray(t *testing.T) {
	ary := js.Global().Get("Array").New("__test__/example.rego", "__test__/example2.rego")
	args := []js.Value{
		ary,
	}

	json, err := outcome(inspect(stubThat, args))
	assert.NoError(t, err)

	cupaloy.SnapshotT(t, json)
}

func TestInspectVinylStream(t *testing.T) {
	stream := js.ValueOf(map[string]any{
		"pipe": js.FuncOf(func(this js.Value, args []js.Value) any {
			handler := args[0]

			handler.Call("write", js.ValueOf(map[string]any{
				"path": "example.rego",
				"contents": js.ValueOf(map[string]any{
					"toString": js.FuncOf(func(this js.Value, args []js.Value) any {
						return rego
					}),
				}),
			}))

			handler.Call("end")

			return nil
		}),
	})

	args := []js.Value{
		stream,
	}

	json, err := outcome(inspect(stubThat, args))
	assert.NoError(t, err)

	cupaloy.SnapshotT(t, json)
}

func TestInspectArrayOfVinyl(t *testing.T) {
	example := js.ValueOf(map[string]any{
		"path":     js.ValueOf("__test__/example.rego"),
		"contents": js.Global().Get("Buffer").Call("from", rego),
	})

	ary := js.Global().Get("Array").New(example)

	args := []js.Value{
		ary,
	}

	json, err := outcome(inspect(stubThat, args))
	assert.NoError(t, err)

	cupaloy.SnapshotT(t, json)
}
