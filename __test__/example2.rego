package example2

# METADATA
# title: Example 2
# description: Second example
# custom:
#   short_name: example2
deny contains msg if {
    msg := "nope"
}
