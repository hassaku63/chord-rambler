package query

import "fmt"

// Basis is the kind of knowledge a fact rests on (根拠の種類 in
// docs/glossary.md). The text says a fact in the words its basis allows
// (docs/architecture/explanations.md). The zero value is Definition.
type Basis int

// The bases the queries return so far.
const (
	// Definition is what the materials define, such as the members of a
	// chord.
	Definition Basis = iota
	// Computation is a fact computed without a convention, such as whether a
	// note is in the scale of the key.
	Computation
	// Convention is a rule of the conventions the readings follow
	// (docs/architecture/theory-conventions.md), such as the kind or the
	// function of a chord.
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
	case Convention:
		return "Convention"
	default:
		return fmt.Sprintf("Basis(%d)", int(b))
	}
}
