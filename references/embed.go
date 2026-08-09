// Package references embeds the note contract so the forge binary is self-contained.
//
// references/schema.yaml is the single source of truth for `forge validate`. It lives
// here rather than under pkg/ because it is a human-facing reference document that
// happens to also be machine-read; go:embed cannot reach a parent directory, so the
// embedding declaration comes to the file instead of the file moving to the code.
package references

import _ "embed"

//go:embed schema.yaml
var SchemaYAML []byte
