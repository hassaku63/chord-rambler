// Package relation computes facts that follow from the definitions whatever
// the school of harmony: the position of a chord seen from the tonic, scale
// membership, the parallel modes that contain a chord, leading tone and
// tritone, root motion and common tones between neighbours, the shape of a
// secondary dominant, and the place of a slot in its bar and phrase.
//
// Part of the engine (docs/architecture/engine.md, "関係の計算"): pure functions,
// no I/O. It may import: material, progression.
package relation
