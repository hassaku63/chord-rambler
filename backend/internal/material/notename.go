package material

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidLetter reports a letter outside C through B.
var ErrInvalidLetter = errors.New("letter is not one of C D E F G A B")

// ErrAlterOutOfRange reports an alteration beyond a double flat or a double
// sharp, whether asked for directly or reached by transposition.
var ErrAlterOutOfRange = errors.New("alteration is beyond a double flat or a double sharp")

// The alteration of a note name runs from a double flat to a double sharp.
const (
	minAlter = -2
	maxAlter = 2
)

// Letter is a natural note name, C through B, in that order.
type Letter int

// The seven letters. C is the zero value.
const (
	C Letter = iota
	D
	E
	F
	G
	A
	B
)

const letters = 7

// naturalSemitones holds the semitones from C up to each letter.
var naturalSemitones = [letters]int{0, 2, 4, 5, 7, 9, 11}

func (l Letter) valid() bool { return l >= C && l <= B }

// String returns the letter as written, such as "F".
func (l Letter) String() string {
	if !l.valid() {
		return fmt.Sprintf("Letter(%d)", int(l))
	}
	return string("CDEFGAB"[l])
}

// NoteName is a spelled note name without an octave: a letter and an
// alteration in semitones (-2 for a double flat up to +2 for a double sharp).
// The zero value is C. Build other values with NewNoteName.
type NoteName struct {
	letter Letter
	alter  int
}

// NewNoteName returns the note name with the letter and the alteration. It
// fails for a letter outside C through B or an alteration outside -2..+2.
func NewNoteName(letter Letter, alter int) (NoteName, error) {
	if !letter.valid() {
		return NoteName{}, fmt.Errorf("note name %s: %w", letter, ErrInvalidLetter)
	}
	if alter < minAlter || alter > maxAlter {
		return NoteName{}, fmt.Errorf("note name %s with alteration %d: %w", letter, alter, ErrAlterOutOfRange)
	}
	return NoteName{letter: letter, alter: alter}, nil
}

// Letter returns the natural note name.
func (n NoteName) Letter() Letter { return n.letter }

// Alter returns the alteration in semitones.
func (n NoteName) Alter() int { return n.alter }

// PitchClass returns the pitch class of the note name (C = 0).
func (n NoteName) PitchClass() PitchClass {
	return PitchClass{v: mod(n.semitones(), semitonesPerOctave)}
}

// String returns the note name for reading, such as "F♯" or "B♭♭". It is not
// the notation of the product, which belongs to the notation package.
func (n NoteName) String() string {
	sign := "♯"
	if n.alter < 0 {
		sign = "♭"
	}
	return n.letter.String() + strings.Repeat(sign, abs(n.alter))
}

// semitones returns the semitones from C up to the note name, without
// reducing modulo 12 (C♭ is -1).
func (n NoteName) semitones() int {
	return naturalSemitones[n.letter] + n.alter
}

// naturalAt returns the semitones from C up to the natural note that is steps
// letters above C. Steps may be negative or reach past B; each seven steps is
// an octave (12 semitones).
func naturalAt(steps int) int {
	octaves := floorDiv(steps, letters)
	return octaves*semitonesPerOctave + naturalSemitones[steps-octaves*letters]
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

func mod(a, b int) int { return a - floorDiv(a, b)*b }

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
