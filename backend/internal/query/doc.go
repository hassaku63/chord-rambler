// Package query answers the questions the screen asks by combining the
// packages below it: candidates for a position, variations of a chord, the
// explanation of one chord, a summary of the progression and role labels. It
// returns structured facts, each with the kind of knowledge it rests on.
//
// Part of the engine (docs/architecture/engine.md, "問い合わせ"): pure functions,
// no I/O. It may import: material, notation, progression, relation, reading, catalog.
package query
