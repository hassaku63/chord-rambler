// Package progression is the data every other engine package works on: a key,
// a meter, and a nesting of section, phrases and slots. A slot holds an
// absolute chord (or no chord, or none chosen yet), a length in beats and an
// optional role.
//
// Part of the engine (docs/architecture/engine.md, "土台: 進行のモデル"): pure functions,
// no I/O. It may import: material.
package progression
