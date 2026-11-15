package material

import (
	"errors"
	"fmt"
)

// ErrDegree reports a degree outside 1 to 7.
var ErrDegree = errors.New("degree is not 1 to 7")

// ErrAlteration reports an alteration of a degree other than a flat, none or
// a sharp.
var ErrAlteration = errors.New("alteration of a degree is not a flat, none or a sharp")

// ErrTarget reports a target that a position does not take: anything other
// than a degree 2 to 6 without alteration (ii, iii, IV, V, vi).
var ErrTarget = errors.New("target is not one of ii, iii, IV, V and vi")

// Alteration is how far the root of a position is moved from the degree of
// the major scale: a flat, none or a sharp. Build positions with NewPosition,
// which checks it.
type Alteration int

// The three alterations of a degree.
const (
	Flat    Alteration = -1
	Natural Alteration = 0
	Sharp   Alteration = 1
)

// Shape is a chord without its root: the intervals of the members above the
// root and an optional bass. The zero value is the root alone. Build other
// values with ShapeOf.
type Shape struct {
	onC Chord // the shape on C
}

// ShapeOf returns the shape of the chord: its intervals and its bass, without
// the root.
func ShapeOf(c Chord) Shape {
	// From C, every interval reaches at most a sharp or a double flat, so these
	// cannot fail.
	onC, _ := NewChord(NoteName{}, c.Intervals()[1:]...)
	if bass, ok := c.Bass(); ok {
		onC, _ = onC.WithBass(bass)
	}
	return Shape{onC: onC}
}

// Intervals returns the intervals of the members above the root, as
// Chord.Intervals does: the perfect unison first.
func (s Shape) Intervals() []Interval { return s.onC.Intervals() }

// Bass returns the bass as an interval above the root, and whether the shape
// has one.
func (s Shape) Bass() (Interval, bool) { return s.onC.Bass() }

// On returns the chord of the shape on the root. It fails when a member or the
// bass would need more than a double sharp or a double flat.
func (s Shape) On(root NoteName) (Chord, error) {
	c, err := NewChord(root, s.Intervals()[1:]...)
	if err != nil {
		return Chord{}, err
	}
	if bass, ok := s.Bass(); ok {
		return c.WithBass(bass)
	}
	return c, nil
}

// Position is a chord seen from the tonic of a major key: a degree of the
// major scale (1 to 7), its alteration, the shape of the chord, and, for a
// secondary chord, the degree it leads to (the /V of V7/V). It has no key; a
// tonic turns it into a chord (Chord). The zero value is degree 1 with the
// shape of the root alone. Build other values with NewPosition.
type Position struct {
	index  int // the degree minus one, so that the zero value is degree 1
	alter  Alteration
	shape  Shape
	target int // the target degree, or 0 for none
}

// NewPosition returns the position of the shape on the degree, moved by the
// alteration. It fails for a degree outside 1 to 7 (ErrDegree) and an
// alteration other than Flat, Natural and Sharp (ErrAlteration).
func NewPosition(degree int, alter Alteration, shape Shape) (Position, error) {
	if degree < 1 || degree > 7 {
		return Position{}, fmt.Errorf("degree %d: %w", degree, ErrDegree)
	}
	if alter < Flat || alter > Sharp {
		return Position{}, fmt.Errorf("alteration %d: %w", int(alter), ErrAlteration)
	}
	return Position{index: degree - 1, alter: alter, shape: shape}, nil
}

// WithTarget returns the position leading to the degree, as V7/V leads to V.
// The degree before "/" is then counted from the root of the target. The
// target is a degree 2 to 6, without alteration; it fails otherwise
// (ErrTarget).
func (p Position) WithTarget(degree int) (Position, error) {
	if degree < 2 || degree > 6 {
		return Position{}, fmt.Errorf("target degree %d: %w", degree, ErrTarget)
	}
	p.target = degree
	return p, nil
}

// Degree returns the degree, 1 to 7.
func (p Position) Degree() int { return p.index + 1 }

// Alteration returns the alteration of the degree.
func (p Position) Alteration() Alteration { return p.alter }

// Shape returns the shape of the chord.
func (p Position) Shape() Shape { return p.shape }

// Target returns the degree the position leads to, and whether it has one.
func (p Position) Target() (int, bool) { return p.target, p.target != 0 }

// interval builds an interval for the tables of this file, which hold only
// valid values.
func interval(number int, q Quality) Interval {
	i, err := NewInterval(number, q)
	if err != nil {
		panic(err) // the tables below hold only valid intervals
	}
	return i
}

// majorIntervals holds the intervals from the tonic up to each degree of the
// major scale.
var majorIntervals = [7]Interval{
	interval(1, Perfect), interval(2, Major), interval(3, Major), interval(4, Perfect),
	interval(5, Perfect), interval(6, Major), interval(7, Major),
}

// Chord returns the chord of the position in the major key on the tonic. Write
// its name with FormatChord, which keeps the spelling (♭VII on D♭ is C♭). The
// root is the note of the degree in the major scale, moved by the alteration
// with its letter kept (♭VII on C is B♭), so the spelling follows
// transposition. With a target, the degree is counted from the root of the
// target (V7/V on C is D7). It fails when the root, a member or the bass would
// need more than a double sharp or a double flat.
func (p Position) Chord(tonic NoteName) (Chord, error) {
	if p.target != 0 {
		t, err := degreeNote(tonic, p.target, Natural)
		if err != nil {
			return Chord{}, err
		}
		tonic = t
	}
	root, err := degreeNote(tonic, p.Degree(), p.alter)
	if err != nil {
		return Chord{}, err
	}
	return p.shape.On(root)
}

// degreeNote returns the note of the degree of the major scale on the tonic,
// moved by the alteration.
func degreeNote(tonic NoteName, degree int, alter Alteration) (NoteName, error) {
	step := majorIntervals[degree-1]
	q := step.Quality()
	switch {
	case alter == Sharp:
		q = Augmented
	case alter == Flat && q == Perfect:
		q = Diminished
	case alter == Flat:
		q = Minor
	}
	// Every number takes augmented, and diminished or minor as chosen above.
	moved, _ := NewInterval(step.Number(), q)
	return tonic.Transpose(moved)
}
