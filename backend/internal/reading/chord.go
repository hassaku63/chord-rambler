package reading

import (
	"fmt"
	"slices"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// Kind is what a reading calls a chord. The zero value, FactsOnly, gives no
// name.
type Kind int

// The kinds of a chord in a major key (docs/architecture/theory-conventions.md,
// 1 to 3).
const (
	// FactsOnly gives no name: only the facts (the position, if any, and the
	// members) are said. A chord spelled enharmonically, such as B♯ major in
	// C, is read so (#89).
	FactsOnly Kind = iota
	// Diatonic is a chord whose notes judged (relation.JudgedNotes) are all
	// notes of the scale of the key, by spelling. Tensions and other basses
	// are not looked at, so G7(♭9) and C/F♯ are diatonic in C; the fourth of
	// Fsus4 (B♭) is, so it is not.
	Diatonic
	// Borrowed is a chord whose notes judged are all notes of a mode on the
	// same tonic other than the major scale, by spelling. Secondary dominants
	// are not borrowed.
	Borrowed
	// SecondaryDominant is a chord of the shape of a secondary dominant
	// (relation.SecondaryDominantShape). It keeps the name whatever follows.
	SecondaryDominant
)

// String returns the name of the constant, for tests and logs. The words
// shown to people belong to the text.
func (k Kind) String() string {
	switch k {
	case FactsOnly:
		return "FactsOnly"
	case Diatonic:
		return "Diatonic"
	case Borrowed:
		return "Borrowed"
	case SecondaryDominant:
		return "SecondaryDominant"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

// Function is the harmonic function of a diatonic chord. The zero value,
// NoFunction, is what every other kind has.
type Function int

// The three functions (docs/architecture/theory-conventions.md, 1).
const (
	NoFunction Function = iota
	Tonic
	Subdominant
	Dominant
)

// String returns the name of the constant, for tests and logs.
func (f Function) String() string {
	switch f {
	case NoFunction:
		return "NoFunction"
	case Tonic:
		return "Tonic"
	case Subdominant:
		return "Subdominant"
	case Dominant:
		return "Dominant"
	default:
		return fmt.Sprintf("Function(%d)", int(f))
	}
}

// functions is the function of each degree of a major key, from 1: I, iii
// and vi are tonic, ii and IV subdominant, V and vii° dominant.
var functions = [...]Function{Tonic, Subdominant, Tonic, Subdominant, Dominant, Tonic, Dominant}

// Reading is one reading of a chord in a key.
type Reading struct {
	// Key is the key the chord was read in.
	Key material.Scale
	// Kind is what the chord is called.
	Kind Kind
	// Function is set for a diatonic chord only, from its degree.
	Function Function
	// Target is the degree a secondary dominant points to (2 to 6), and 0
	// for the other kinds.
	Target int
	// Modes are the parallel modes that hold the notes judged of a borrowed
	// chord by spelling, in the order of material.Mode, and nil for the other
	// kinds.
	Modes []material.Mode
}

// FromParallelMinor reports whether a borrowed chord is said to come from the
// parallel minor: its modes hold the natural minor (Aeolian). When it does
// not, the chord is said to come from the modes by name.
func (r Reading) FromParallelMinor() bool {
	return r.Kind == Borrowed && slices.Contains(r.Modes, material.Aeolian)
}

// ChordReading holds the readings of one chord: the main reading, and the
// others that may be listed beside it. For now there are no others (#101,
// B1).
type ChordReading struct {
	Main   Reading
	Others []Reading
}

// ReadChord reads the chord in the major key. The kind is decided in this
// order:
//
//  1. a secondary dominant, when the chord has the shape of one
//     (relation.SecondaryDominantShape);
//  2. diatonic, when the scale of the key holds the notes judged
//     (relation.JudgedNotes: the root, the second to the seventh members and
//     a bass a minor seventh above the root) by spelling;
//  3. borrowed, when another mode on the tonic holds them by spelling;
//  4. facts only, otherwise. A chord spelled enharmonically (A♯ or B♯ major
//     in C) is read so, and so is a chord whose root has no position.
//
// A note is held by a mode when the interval from the tonic up to it is one
// of the mode, so a mode that cannot be spelled on the tonic within a double
// sharp or flat is compared as well.
//
// It fails for a key other than Ionian (relation.ErrNotMajor).
func ReadChord(c material.Chord, key material.Scale) (ChordReading, error) {
	main, err := readMain(c, key)
	if err != nil {
		return ChordReading{}, err
	}
	return ChordReading{Main: main}, nil
}

func readMain(c material.Chord, key material.Scale) (Reading, error) {
	r := Reading{Key: key}
	target, ok, err := relation.SecondaryDominantShape(c, key)
	if err != nil {
		return Reading{}, fmt.Errorf("read chord: %w", err)
	}
	if ok {
		r.Kind, r.Target = SecondaryDominant, target
		return r, nil
	}
	modes := modesHolding(relation.JudgedNotes(c), key.Tonic())
	if slices.Contains(modes, material.Ionian) {
		pos, err := relation.PositionOf(c, key)
		if err != nil {
			// Not reached: the key is Ionian (SecondaryDominantShape would
			// have failed) and the root is a note of its scale, so it has a
			// position without alteration.
			panic(err)
		}
		r.Kind, r.Function = Diatonic, functions[pos.Degree()-1]
		return r, nil
	}
	if len(modes) > 0 {
		r.Kind, r.Modes = Borrowed, modes
	}
	return r, nil
}

// modesHolding returns the modes on the tonic, in the order of material.Mode,
// that hold every note by spelling: the interval from the tonic up to the note
// is the interval from C up to a note of the mode on C.
func modesHolding(notes []material.NoteName, tonic material.NoteName) []material.Mode {
	var out []material.Mode
	for m := material.Ionian; m <= material.Locrian; m++ {
		// Every mode on C is spelled within a single sharp or flat.
		onC, _ := material.NewScale(material.NoteName{}, m)
		var degrees []material.Interval
		for _, n := range onC.Notes() {
			// Two notes of a mode on C always have an interval.
			iv, _ := material.IntervalBetween(material.NoteName{}, n)
			degrees = append(degrees, iv)
		}
		if holdsAll(degrees, notes, tonic) {
			out = append(out, m)
		}
	}
	return out
}

func holdsAll(degrees []material.Interval, notes []material.NoteName, tonic material.NoteName) bool {
	for _, n := range notes {
		// No interval of the five qualities is no interval of a mode either.
		iv, err := material.IntervalBetween(tonic, n)
		if err != nil || !slices.Contains(degrees, iv) {
			return false
		}
	}
	return true
}
