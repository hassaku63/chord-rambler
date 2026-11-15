// Package material defines the musical materials: spelled note names and
// intervals, pitch-class sets, chords (a root, a set of intervals above it and an
// optional bass), scales, transposition that keeps spelling, and positions (a
// chord seen from the tonic of a major key). A chord carries no key, and a
// position carries none either. This is knowledge by definition.
//
// Part of the engine (docs/architecture/engine.md, "音の素材"): pure functions,
// no I/O. It imports no other engine package.
package material
