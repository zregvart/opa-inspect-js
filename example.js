const opa = require("./index");

opa.inspect(
    "example.rego",
    `package hmm

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
  console.log(JSON.stringify(json));
});
