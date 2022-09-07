package main

import (
	"syscall/js"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestInspectProvidedFile(t *testing.T) {
	that := js.Null()
	args := []js.Value{
		js.ValueOf("example.rego"),
		js.ValueOf(`package hmm

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
`),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoaded(t *testing.T) {
	that := js.Null()
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedSecondArgumentNull(t *testing.T) {
	that := js.Null()
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
		js.Null(),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}

func TestInspectSingleFileLoadedSecondArgumentUndefined(t *testing.T) {
	that := js.Null()
	args := []js.Value{
		js.ValueOf("__test__/example.rego"),
		js.Undefined(),
	}

	json := inspect(that, args)

	cupaloy.SnapshotT(t, json)
}
