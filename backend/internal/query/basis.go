package query

import (
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
)

// Basis is the kind of knowledge a fact rests on (根拠の種類 in
// docs/glossary.md). The text says a fact in the words its basis allows
// (docs/architecture/explanations.md). The zero value is none of them: a
// label that never set its basis shows as Basis(0).
type Basis int

// The bases the queries return so far.
const (
	// Definition is what the materials define, such as the members of a
	// chord.
	Definition Basis = iota + 1
	// Computation is a fact computed without a rule of a school, such as whether a
	// note is in the scale of the key.
	Computation
	// School is a rule of the school of theory the readings follow (流儀,
	// docs/architecture/theory-conventions.md), such as the kind or the
	// function of a chord. It is not a convention of practice (慣習), which
	// the catalogues hold.
	School
	// Convention is a preference of a style that a catalogue holds (慣習),
	// such as how often one function moves to another. It is said as a
	// frequency, never as right or wrong.
	Convention
)

// String returns the name of the constant, for tests and logs. The words
// shown to people belong to the text.
func (b Basis) String() string {
	switch b {
	case Definition:
		return "Definition"
	case Computation:
		return "Computation"
	case School:
		return "School"
	case Convention:
		return "Convention"
	default:
		return fmt.Sprintf("Basis(%d)", int(b))
	}
}

// basisOfGrounds returns the basis of an entry of a catalogue from its
// grounds. The switch lists every constant of catalog.Grounds and has no
// default, so the lint (exhaustive) fails when a constant is added. The zero
// value (grounds never set) and any other value that is none of the
// constants map to Basis(0).
func basisOfGrounds(g catalog.Grounds) Basis {
	switch g {
	case catalog.Convention:
		return Convention
	case 0:
		return 0
	}
	return 0
}
