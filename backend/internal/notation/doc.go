// Package notation reads and writes chord names and Roman numerals. Chord names
// are read with their aliases, and may be written with the display spelling
// when shown without a key; Roman numerals are read in the display form only.
// These are conventions that can be switched.
//
// Part of the engine (docs/architecture/engine.md, "表記"): pure functions,
// no I/O. It may import: material.
package notation
