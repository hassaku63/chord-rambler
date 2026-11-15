package material

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrPitchClassOutOfRange reports a pitch class outside 0..11.
var ErrPitchClassOutOfRange = errors.New("pitch class is outside 0..11")

const semitonesPerOctave = 12

// PitchClass is a pitch without spelling or octave, 0 to 11 with C = 0. C♯
// and D♭ are the same pitch class. The zero value is C.
type PitchClass struct {
	v int
}

// NewPitchClass returns the pitch class n. It fails outside 0..11.
func NewPitchClass(n int) (PitchClass, error) {
	if n < 0 || n >= semitonesPerOctave {
		return PitchClass{}, fmt.Errorf("pitch class %d: %w", n, ErrPitchClassOutOfRange)
	}
	return PitchClass{v: n}, nil
}

// Int returns the pitch class as a number, 0 to 11.
func (pc PitchClass) Int() int { return pc.v }

// String returns the number, such as "4".
func (pc PitchClass) String() string { return strconv.Itoa(pc.v) }

// PitchClassSet is a set of the twelve pitch classes. It has no spelling. The
// zero value is the empty set.
type PitchClassSet struct {
	bits uint16
}

// NewPitchClassSet returns the set of the pitch classes. Repeats count once.
func NewPitchClassSet(pcs ...PitchClass) PitchClassSet {
	var s PitchClassSet
	for _, pc := range pcs {
		s.bits |= 1 << pc.v
	}
	return s
}

// Contains reports whether the set has the pitch class.
func (s PitchClassSet) Contains(pc PitchClass) bool { return s.bits&(1<<pc.v) != 0 }

// IsSubsetOf reports whether every pitch class of s is in other.
func (s PitchClassSet) IsSubsetOf(other PitchClassSet) bool { return s.bits&^other.bits == 0 }

// Union returns the pitch classes in s, in other, or in both.
func (s PitchClassSet) Union(other PitchClassSet) PitchClassSet {
	return PitchClassSet{bits: s.bits | other.bits}
}

// Transpose returns the set with every pitch class moved up by the semitones,
// modulo 12. Negative semitones move down.
func (s PitchClassSet) Transpose(semitones int) PitchClassSet {
	var out PitchClassSet
	for _, pc := range s.PitchClasses() {
		out.bits |= 1 << mod(pc.v+semitones, semitonesPerOctave)
	}
	return out
}

// PitchClasses returns the members in ascending order.
func (s PitchClassSet) PitchClasses() []PitchClass {
	var pcs []PitchClass
	for v := range semitonesPerOctave {
		if s.bits&(1<<v) != 0 {
			pcs = append(pcs, PitchClass{v: v})
		}
	}
	return pcs
}

// String returns the members for reading, such as "{0 4 7}".
func (s PitchClassSet) String() string {
	parts := make([]string, 0, semitonesPerOctave)
	for _, pc := range s.PitchClasses() {
		parts = append(parts, pc.String())
	}
	return "{" + strings.Join(parts, " ") + "}"
}
