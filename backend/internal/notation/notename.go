package notation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrNoteName reports text that does not start with a note name: an uppercase
// letter A to G.
var ErrNoteName = errors.New("not a note name: an uppercase letter A to G")

// ErrAccidental reports accidentals beyond the accepted spellings: one of ♯ #
// ♭ b, or a double one, 𝄪 x ## ♯♯ or 𝄫 bb ♭♭.
var ErrAccidental = errors.New("accidental is beyond a double sharp or flat, or not an accepted spelling")

// ErrTrailing reports text left after a whole name has been read.
var ErrTrailing = errors.New("unexpected text after the name")

// ParseError reports where a name could not be read. Rest is the unread part
// of Input, starting where reading stopped; Err says why. Err is one of the
// Err values of this package, or wraps an error of the material package (a
// member beyond a double sharp, a bass with no interval to the root).
type ParseError struct {
	Input string
	Rest  string
	Err   error
}

func (e *ParseError) Error() string {
	if e.Rest == "" {
		return fmt.Sprintf("cannot read %q: it ends early: %v", e.Input, e.Err)
	}
	return fmt.Sprintf("cannot read %q in %q: %v", e.Rest, e.Input, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// accidentals holds the accepted spellings of the alterations. They are tried
// in order, so each double spelling comes before the single one it starts with
// ("bb" before "b", "♯♯" before "♯").
var accidentals = []struct {
	text  string
	alter int
}{
	{"##", 2}, {"♯♯", 2}, {"bb", -2}, {"♭♭", -2},
	{"𝄪", 2}, {"x", 2}, {"𝄫", -2},
	{"♯", 1}, {"#", 1}, {"♭", -1}, {"b", -1},
}

// written holds the accidental written for each alteration, from -2 to +2.
var written = [...]string{"𝄫", "♭", "", "♯", "𝄪"}

// ParseNoteName reads a note name: an uppercase letter and at most one
// accidental (♯ or #, ♭ or b, 𝄪 x ## or ♯♯ for a double sharp, 𝄫 bb or ♭♭ for a
// double flat). It fails with a *ParseError.
func ParseNoteName(s string) (material.NoteName, error) {
	n, rest, err := readNoteName(s)
	if err != nil {
		return material.NoteName{}, &ParseError{Input: s, Rest: rest, Err: err}
	}
	if rest != "" {
		return material.NoteName{}, &ParseError{Input: s, Rest: rest, Err: ErrTrailing}
	}
	return n, nil
}

// FormatNoteName writes the note name with ♯ and ♭, and 𝄪 and 𝄫 for the double
// ones, keeping the spelling (B𝄫 stays B𝄫). DisplaySpelling respells it first.
func FormatNoteName(n material.NoteName) string {
	return n.Letter().String() + written[n.Alter()+2]
}

// readNoteName reads a note name at the start of s and returns the rest. On
// failure, rest is where reading stopped.
func readNoteName(s string) (material.NoteName, string, error) {
	if s == "" || s[0] < 'A' || s[0] > 'G' {
		return material.NoteName{}, s, ErrNoteName
	}
	letter := material.Letter((int(s[0]-'A') + 5) % 7) // A is the sixth letter from C
	rest, alter := s[1:], 0
	for _, a := range accidentals {
		if strings.HasPrefix(rest, a.text) {
			rest, alter = rest[len(a.text):], a.alter
			break
		}
	}
	for _, a := range accidentals {
		if strings.HasPrefix(rest, a.text) {
			return material.NoteName{}, s[1:], ErrAccidental
		}
	}
	// The letter and the alteration are in range, so this cannot fail.
	n, err := material.NewNoteName(letter, alter)
	return n, rest, err
}

// DisplaySpelling returns the note name to show: a double sharp or a double
// flat, C♭ and F♭ become the neighbouring letter that sounds the same (B♭♭ is
// A, C♭ is B, F♭ is E, B𝄪 is C♯). Other note names come back as they are
// (docs/architecture/theory-conventions.md, 6). Only the shown text changes;
// the engine keeps its own spelling. It is only for a chord name shown without
// a key; chord names in a key keep their spelling (FormatChord).
func DisplaySpelling(n material.NoteName) material.NoteName {
	// A diminished second keeps the pitch and moves the letter by one.
	d2, _ := material.NewInterval(2, material.Diminished)
	var out material.NoteName
	switch {
	case n.Alter() == 2:
		out, _ = n.Transpose(d2)
	case n.Alter() == -2, n.Alter() == -1 && (n.Letter() == material.C || n.Letter() == material.F):
		out, _ = n.TransposeDown(d2)
	default:
		return n
	}
	// Both moves end on a natural or a single sharp or flat, so they cannot fail.
	return out
}
