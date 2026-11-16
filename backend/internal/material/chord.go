package material

import (
	"errors"
	"fmt"
)

// ErrDuplicateNumber reports a chord given two intervals of the same number,
// such as a major and a minor third. The root is the unison, so a unison
// given as a member is a duplicate too.
var ErrDuplicateNumber = errors.New("chord has two intervals of the same number")

// ErrOctaveMember reports a chord member or bass that repeats another number
// an octave higher: an 8th, 10th or 12th as a member, or any compound interval
// as the bass. Such a member would sound a third or a fifth that the views do
// not see (a sus4 with a tenth sounds a third).
var ErrOctaveMember = errors.New("chord member repeats a number an octave higher")

// Chord is a chord without a key: a root, the intervals of the other members
// above the root (at most one per interval number), and an optional bass as an
// interval above the root. The root is always a member, as the unison. Kinds
// and views (major triad, minor seventh, tensions) are derived from the
// intervals, not stored. The zero value is C alone. Build other values with
// NewChord.
type Chord struct {
	root    NoteName
	members [maxNumber]member // indexed by number - 1; index 0 (the root) is unused
	bass    member
}

type member struct {
	present  bool
	interval Interval
}

// NewChord returns the chord on the root with the members the intervals above
// it. Members are 2nds to 7ths and the tensions 9, 11 and 13. The order of the
// intervals does not matter. It fails for two intervals of the same number
// (ErrDuplicateNumber), for an 8th, 10th or 12th (ErrOctaveMember) and for a
// member that would need more than a double sharp or a double flat
// (ErrAlterOutOfRange).
func NewChord(root NoteName, intervals ...Interval) (Chord, error) {
	c := Chord{root: root}
	for _, iv := range intervals {
		if iv.steps == 0 || c.members[iv.steps].present {
			return Chord{}, fmt.Errorf("chord on %s with %s: %w", root, iv, ErrDuplicateNumber)
		}
		if n := iv.Number(); n == 8 || n == 10 || n == 12 {
			return Chord{}, fmt.Errorf("chord on %s with %s: %w", root, iv, ErrOctaveMember)
		}
		if _, err := root.Transpose(iv); err != nil {
			return Chord{}, fmt.Errorf("chord on %s: %w", root, err)
		}
		c.members[iv.steps] = member{present: true, interval: iv}
	}
	return c, nil
}

// WithBass returns the chord with the bass the interval above the root, a
// simple interval (1 to 7). The bass need not be a member. It fails for a
// compound interval (ErrOctaveMember) and when the bass would need more than a
// double sharp or a double flat.
func (c Chord) WithBass(bass Interval) (Chord, error) {
	if bass.Number() > 7 {
		return Chord{}, fmt.Errorf("bass %s of the chord on %s: %w", bass, c.root, ErrOctaveMember)
	}
	if _, err := c.root.Transpose(bass); err != nil {
		return Chord{}, fmt.Errorf("bass of the chord on %s: %w", c.root, err)
	}
	c.bass = member{present: true, interval: bass}
	return c, nil
}

// Root returns the root.
func (c Chord) Root() NoteName { return c.root }

// Bass returns the bass as an interval above the root, and whether the chord
// has one.
func (c Chord) Bass() (Interval, bool) { return c.bass.interval, c.bass.present }

// Intervals returns the intervals of the members above the root in ascending
// number, starting with the perfect unison of the root.
func (c Chord) Intervals() []Interval {
	ivs := make([]Interval, 1, maxNumber)
	for _, m := range c.members {
		if m.present {
			ivs = append(ivs, m.interval)
		}
	}
	return ivs
}

// Notes returns the spelled members in the order of Intervals: the root first.
func (c Chord) Notes() []NoteName {
	ivs := c.Intervals()
	notes := make([]NoteName, 0, len(ivs))
	for _, iv := range ivs {
		// NewChord checked that every member is in range.
		n, _ := c.root.Transpose(iv)
		notes = append(notes, n)
	}
	return notes
}

// PitchClasses returns the pitch classes of the members. A bass that is not a
// member is not included.
func (c Chord) PitchClasses() PitchClassSet {
	var s PitchClassSet
	for _, n := range c.Notes() {
		s = s.Union(NewPitchClassSet(n.PitchClass()))
	}
	return s
}

// SoundingPitchClasses returns the pitch classes that sound: the members and
// the bass. A comparison by pitch class (the tritone, the common tones of two
// chords) uses this set; a bass that is not a member (C/B♭) can
// change the answer. A comparison with a key or a mode, such as the kind of a
// reading, is by spelling and looks at spelled notes instead.
func (c Chord) SoundingPitchClasses() PitchClassSet {
	s := c.PitchClasses()
	if bass, ok := c.Bass(); ok {
		// WithBass checked that the bass is in range.
		n, _ := c.root.Transpose(bass)
		s = s.Union(NewPitchClassSet(n.PitchClass()))
	}
	return s
}

// Transpose returns the chord with the root moved up by the interval
// (down by a half step for the diminished unison). The intervals and the bass
// stay, so the spelling is kept: F major up a major second is G major, and E♭ minor up a minor second is F♭ minor. It fails
// when the root, a member or the bass would need more than a double sharp or
// a double flat.
func (c Chord) Transpose(iv Interval) (Chord, error) {
	root, err := c.root.Transpose(iv)
	if err != nil {
		return Chord{}, err
	}
	moved, err := NewChord(root, c.Intervals()[1:]...)
	if err != nil {
		return Chord{}, err
	}
	if bass, ok := c.Bass(); ok {
		return moved.WithBass(bass)
	}
	return moved, nil
}

// has returns the member of the number and whether the chord has it.
func (c Chord) has(number int) (Interval, bool) {
	m := c.members[number-1]
	return m.interval, m.present
}
