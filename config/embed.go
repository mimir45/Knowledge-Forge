// Package config carries the packaged configuration templates as compiled-in bytes.
//
// It holds no logic — pkg/config does the loading. It exists as its own package for one
// reason: go:embed cannot reach outside its own directory, so these
// files live at the repo root under config/, where a reader looks for them. Embedding rather
// than reading from disk means a stranger who unpacks only the binary still has a
// complete base layer; the rule that the packaged layer is never user-edited is then
// enforced by construction, because there is nothing on disk to edit.
package config

import (
	"embed"
	_ "embed"
)

// Example is the lowest layer of the chain: the schema is the union of ADDENDUM §E and
// the DESIGN §10 keys §E never restates.
//
//go:embed forge.config.example.md
var Example []byte

// Presets are the overlays `forge init` applies on top of Example. Two independent axes:
// engine presets (offline, claude-only, byo-api, max) decide what may make model calls,
// stack presets (java-backend, frontend, devops, minimal) decide what the static core
// indexes. One of each is a normal answer.
//
//go:embed presets/*.md
var Presets embed.FS
