// Package reading names what relation computes, following the conventions
// recorded with their sources in docs/architecture/theory-conventions.md:
// function, kind (diatonic, borrowed, secondary dominant), cadence,
// resolution and the handling of tensions. A reading is never stored; it is
// derived from the whole progression and carries the key it was read in.
//
// Part of the engine (docs/architecture/engine.md, "読み"): pure functions,
// no I/O. It may import: material, progression, relation.
package reading
