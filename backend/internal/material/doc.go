// Package material defines the musical materials: spelled note names and
// intervals, pitch-class sets, chords (a root, a set of degrees above it and an
// optional bass), scales, and transposition that keeps spelling. A chord
// carries no key. This is knowledge by definition.
//
// Part of the engine (docs/architecture/engine.md, "音の素材"): pure functions,
// no I/O. It imports no other engine package.
package material
