package main

import (
	"syscall/js"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

var that = js.ValueOf(map[string]any{})

var rego = `package example

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
}
`

func TestInspectProvidedFile(t *testing.T) {
	args := []js.Value{
		js.ValueOf("example.rego"),
		js.ValueOf(rego),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoaded(t *testing.T) {
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedSecondArgumentNull(t *testing.T) {
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
		js.Null(),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedSecondArgumentUndefined(t *testing.T) {
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
		js.Undefined(),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedViaCustomReadFunction(t *testing.T) {
	read := js.FuncOf(func(this js.Value, args []js.Value) any {
		bytes := []byte(rego)
		ary := js.Global().Get("Uint8Array").New(len(bytes))

		js.CopyBytesToJS(ary, bytes)

		return ary
	})

	that := js.ValueOf(map[string]any{
		"read": read,
	})
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
	}

	json := inspect(that, args)

	print(json)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileGivenAsArray(t *testing.T) {
	ary := js.Global().Get("Array").New("__test__/example.rego")
	args := []js.Value{
		ary,
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectMultipleFilesGivenAsArray(t *testing.T) {
	ary := js.Global().Get("Array").New("__test__/example.rego", "__test__/example2.rego")
	args := []js.Value{
		ary,
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}
