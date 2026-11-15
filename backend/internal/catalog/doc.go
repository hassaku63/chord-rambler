// Package catalog loads the catalogues: idiomatic progressions and the
// preferences between chords (conventions), and the weighted links between
// impressions and harmonic means (associations). They are JSON data embedded
// in the binary and validated when loaded.
//
// Part of the engine (docs/architecture/engine.md, "慣習、意図と手段"): pure functions,
// no I/O. It may import: material, notation, progression, relation, reading.
package catalog
