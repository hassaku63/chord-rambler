package material

import (
	"errors"
	"fmt"
)

// ErrInvalidMode reports a mode outside the seven church modes.
var ErrInvalidMode = errors.New("mode is not one of the seven church modes")

// Mode is one of the seven church modes, Ionian through Locrian, in the order
// of their starting letters on the natural notes (C through B). Ionian is the
// zero value.
type Mode int

// The seven modes.
const (
	Ionian Mode = iota // the major scale
	Dorian
	Phrygian
	Lydian
	Mixolydian
	Aeolian // the natural minor scale
	Locrian
)

func (m Mode) valid() bool { return m >= Ionian && m <= Locrian }

// String returns the name of the mode, such as "Dorian".
func (m Mode) String() string {
	names := [...]string{"Ionian", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Aeolian", "Locrian"}
	if !m.valid() {
		return fmt.Sprintf("Mode(%d)", int(m))
	}
	return names[m]
}

// Scale is a tonic and a mode: seven spelled notes, one on each letter. The
// zero value is C Ionian (C major). Build other values with NewScale.
type Scale struct {
	tonic NoteName
	mode  Mode
}

// NewScale returns the scale of the mode on the tonic. It fails for a mode
// outside the seven (ErrInvalidMode) and for a note that would need more than
// a double sharp or a double flat (ErrAlterOutOfRange).
func NewScale(tonic NoteName, mode Mode) (Scale, error) {
	if !mode.valid() {
		return Scale{}, fmt.Errorf("scale on %s: %w", tonic, ErrInvalidMode)
	}
	s := Scale{tonic: tonic, mode: mode}
	if _, err := s.spell(); err != nil {
		return Scale{}, err
	}
	return s, nil
}

// Tonic returns the tonic.
func (s Scale) Tonic() NoteName { return s.tonic }

// Mode returns the mode.
func (s Scale) Mode() Mode { return s.mode }

// Notes returns the seven notes from the tonic upward.
func (s Scale) Notes() []NoteName {
	// NewScale checked that every note is in range.
	notes, _ := s.spell()
	return notes[:]
}

// PitchClasses returns the pitch classes of the seven notes.
func (s Scale) PitchClasses() PitchClassSet {
	var set PitchClassSet
	for _, n := range s.Notes() {
		set = set.Union(NewPitchClassSet(n.PitchClass()))
	}
	return set
}

// Contains reports whether every pitch class of the set is in the scale.
func (s Scale) Contains(set PitchClassSet) bool { return set.IsSubsetOf(s.PitchClasses()) }

// spell returns the notes. A mode is the natural notes read from its starting
// letter (Dorian from D), so the interval from the tonic to each note is the
// interval between those natural notes.
func (s Scale) spell() ([letters]NoteName, error) {
	var notes [letters]NoteName
	start := NoteName{letter: Letter(s.mode)}
	for i := range letters {
		// Two natural notes always have an interval.
		iv, _ := IntervalBetween(start, NoteName{letter: Letter((int(s.mode) + i) % letters)})
		n, err := s.tonic.Transpose(iv)
		if err != nil {
			return notes, fmt.Errorf("%s %s: %w", s.tonic, s.mode, err)
		}
		notes[i] = n
	}
	return notes, nil
}
