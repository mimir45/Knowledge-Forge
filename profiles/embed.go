// Package profiles carries the developer-profile template as compiled-in bytes.
//
// Like the config package it holds no logic and exists only because go:embed cannot
// reach outside its own directory, and DESIGN §9 fixes the template at profiles/ in the
// repo root. `forge init` renders it to <vault>/profiles/me.md — the rendered copy is
// the human-editable one, this is the mould.
package profiles

import _ "embed"

// Template is DESIGN §9's profile, as a text/template. AUDIT §8.4 D-4: `forge init` is
// the only writer of the rendered file, and the forge-init skill shells out to it rather
// than writing a second copy of this shape in prose.
//
//go:embed me.template.md
var Template []byte
