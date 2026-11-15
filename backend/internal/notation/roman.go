package notation

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrDegree reports a degree outside 1 to 7.
var ErrDegree = errors.New("degree is not 1 to 7")

// ErrAlteration reports an alteration of a degree other than a flat, none or
// a sharp.
var ErrAlteration = errors.New("alteration of a degree is not a flat, none or a sharp")

// ErrTarget reports a target that Roman numerals do not write: anything after
// "/" other than a degree 2 to 6 without an accidental, written in the case of
// its triad in the major key (ii, iii, IV, V, vi). Text after the one target,
// as in V7/V/V, is ErrTrailing.
var ErrTarget = errors.New("target is not one of ii, iii, IV, V and vi")

// ErrRomanBass reports a position with a bass. Roman numerals here do not
// write inversions or a bass.
var ErrRomanBass = errors.New("a bass is not written in Roman numerals")

// ErrNumeral reports text that does not start with a Roman numeral, I to VII
// or i to vii, after at most one ♭ or ♯.
var ErrNumeral = errors.New("not a Roman numeral I to VII, after at most one ♭ or ♯")

// ErrRomanKind reports a chord kind after the numeral that is not written in
// that case, such as "iimaj7" or "V-".
var ErrRomanKind = errors.New("unknown chord kind for the case of the numeral")

// ErrRomanCase reports a chord without a third (sus2, sus4, 7sus4) whose
// numeral is not in its case: the case of the triad of its degree in the major
// key, or uppercase on an altered degree. Examples are "vsus4" and "♭viisus4".
var ErrRomanCase = errors.New("a chord without a third takes the case of the triad of its degree, or uppercase on an altered degree")

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
	onC material.Chord // the shape on C
}

// ShapeOf returns the shape of the chord: its intervals and its bass, without
// the root.
func ShapeOf(c material.Chord) Shape {
	// From C, every interval reaches at most a sharp or a double flat, so these
	// cannot fail.
	onC, _ := material.NewChord(material.NoteName{}, c.Intervals()[1:]...)
	if bass, ok := c.Bass(); ok {
		onC, _ = onC.WithBass(bass)
	}
	return Shape{onC: onC}
}

// Intervals returns the intervals of the members above the root, as
// material.Chord.Intervals does: the perfect unison first.
func (s Shape) Intervals() []material.Interval { return s.onC.Intervals() }

// Bass returns the bass as an interval above the root, and whether the shape
// has one.
func (s Shape) Bass() (material.Interval, bool) { return s.onC.Bass() }

// On returns the chord of the shape on the root. It fails when a member or the
// bass would need more than a double sharp or a double flat.
func (s Shape) On(root material.NoteName) (material.Chord, error) {
	c, err := material.NewChord(root, s.Intervals()[1:]...)
	if err != nil {
		return material.Chord{}, err
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

// majorIntervals holds the intervals from the tonic up to each degree of the
// major scale.
var majorIntervals = [7]material.Interval{
	iv(1, material.Perfect), iv(2, material.Major), iv(3, material.Major), iv(4, material.Perfect),
	iv(5, material.Perfect), iv(6, material.Major), iv(7, material.Major),
}

// Chord returns the chord of the position in the major key on the tonic. Write
// its name with FormatChord, which keeps the spelling (♭VII on D♭ is C♭). The
// root is the note of the degree in the major scale, moved by the alteration
// with its letter kept (♭VII on C is B♭), so the spelling follows
// transposition. With a target, the degree is counted from the root of the
// target (V7/V on C is D7). It fails when the root, a member or the bass would
// need more than a double sharp or a double flat.
func (p Position) Chord(tonic material.NoteName) (material.Chord, error) {
	if p.target != 0 {
		t, err := degreeNote(tonic, p.target, Natural)
		if err != nil {
			return material.Chord{}, err
		}
		tonic = t
	}
	root, err := degreeNote(tonic, p.Degree(), p.alter)
	if err != nil {
		return material.Chord{}, err
	}
	return p.shape.On(root)
}

// degreeNote returns the note of the degree of the major scale on the tonic,
// moved by the alteration.
func degreeNote(tonic material.NoteName, degree int, alter Alteration) (material.NoteName, error) {
	step := majorIntervals[degree-1]
	q := step.Quality()
	switch {
	case alter == Sharp:
		q = material.Augmented
	case alter == Flat && q == material.Perfect:
		q = material.Diminished
	case alter == Flat:
		q = material.Minor
	}
	// Every number takes augmented, and diminished or minor as chosen above.
	moved, _ := material.NewInterval(step.Number(), q)
	return tonic.Transpose(moved)
}

// letterCase is how the numeral of a chord kind is written.
type letterCase int

const (
	upper    letterCase = iota // major and augmented triads
	lower                      // minor and diminished triads
	byDegree                   // no third: see noThirdCase
)

// romanForm is how a chord kind is written after the numeral.
type romanForm struct {
	kind int // the index of the kind in kinds
	text string
	c    letterCase
}

// romanForms is the Roman numeral form of each kind of the vocabulary
// (docs/architecture/theory-conventions.md, 6). The owner's decision C1 in
// Issue #78: uppercase for major and augmented, lowercase for minor and
// diminished, °, ø7 and +, and the kinds of chord names without the m. A
// convention of our own: a chord without a third takes its suffix on the
// degree (Vsus4) in the case of noThirdCase (byDegree). Decided in the
// implementation (#80): each kind has one form and no aliases, as only the
// display forms are read. Building the table checks that
// every kind in kinds has exactly one form and every form names a kind in
// kinds, and panics otherwise. Longer texts come first, so that "°7" is tried
// before "°" and "7sus4" before "7".
var romanForms = func() []romanForm {
	byName := []struct {
		kind, text string
		c          letterCase
	}{
		{"", "", upper}, {"m", "", lower}, {"dim", "°", lower}, {"aug", "+", upper},
		{"7", "7", upper}, {"maj7", "maj7", upper}, {"m7", "7", lower},
		{"m7(♭5)", "ø7", lower}, {"dim7", "°7", lower},
		{"sus2", "sus2", byDegree}, {"sus4", "sus4", byDegree}, {"7sus4", "7sus4", byDegree},
		{"add9", "add9", upper},
	}
	forms := make([]romanForm, 0, len(byName))
	count := make([]int, len(kinds))
	for _, f := range byName {
		i := slices.IndexFunc(kinds, func(k kind) bool { return k.name == f.kind })
		if i < 0 {
			panic(fmt.Sprintf("Roman numeral form for %q, which is not a chord kind", f.kind))
		}
		count[i]++
		forms = append(forms, romanForm{kind: i, text: f.text, c: f.c})
	}
	for i, n := range count {
		if n != 1 {
			panic(fmt.Sprintf("chord kind %q has %d Roman numeral forms, want 1", kinds[i].name, n))
		}
	}
	slices.SortStableFunc(forms, func(a, b romanForm) int { return len(b.text) - len(a.text) })
	return forms
}()

// numerals holds the uppercase numerals of the degrees 1 to 7.
var numerals = [7]string{"I", "II", "III", "IV", "V", "VI", "VII"}

// majorKeyCase holds the case of the triad of each degree in the major key:
// I ii iii IV V vi vii°.
var majorKeyCase = [7]letterCase{upper, lower, lower, upper, upper, lower, lower}

// noThirdCase returns the case of a chord without a third on the degree, moved
// by the alteration: the case of the triad of the degree in the major key, or
// uppercase on an altered degree, as the major triad on it (♭VIIsus4; the
// owner's decision in #82).
func noThirdCase(degree int, alter Alteration) letterCase {
	if alter != Natural {
		return upper
	}
	return majorKeyCase[degree-1]
}

func numeral(degree int, c letterCase) string {
	if c == lower {
		return strings.ToLower(numerals[degree-1])
	}
	return numerals[degree-1]
}

// FormatRoman writes the position as a Roman numeral: an accidental for an
// altered degree, the numeral in the case of the chord (uppercase for major
// and augmented, lowercase for minor and diminished, the case of the triad of
// the degree in the major key for a chord without a third, or uppercase on an
// altered degree), the kind and tensions written as in chord names, and "/"
// with the target. Examples are ii7, Imaj7, viiø7, ♭VII, V7(9), Vsus4,
// ♭VIIsus4 and V7/V. A chord name shown with it is written by FormatChord. It
// fails for a position with a bass (ErrRomanBass) and a shape outside the
// vocabulary (ErrNoName).
func FormatRoman(p Position) (string, error) {
	if _, ok := p.shape.Bass(); ok {
		return "", ErrRomanBass
	}
	suffix, err := chordSuffix(p.shape.Intervals()[1:])
	if err != nil {
		return "", fmt.Errorf("shape %v: %w", p.shape.Intervals()[1:], err)
	}
	form, tensions := romanFormOf(suffix)
	c := form.c
	if c == byDegree {
		c = noThirdCase(p.Degree(), p.alter)
	}
	s := [...]string{"♭", "", "♯"}[p.alter+1] + numeral(p.Degree(), c) + form.text + tensions
	if p.target != 0 {
		s += "/" + numeral(p.target, majorKeyCase[p.target-1])
	}
	return s, nil
}

// romanFormOf returns the form of the kind of a chord suffix, and the tensions
// in parentheses that follow the kind.
func romanFormOf(suffix string) (romanForm, string) {
	if i := slices.IndexFunc(romanForms, func(f romanForm) bool { return kinds[f.kind].name == suffix }); i >= 0 {
		return romanForms[i], ""
	}
	// Only 7, maj7 and m7 take tensions, and none of them is the start of
	// another with "(".
	for _, f := range romanForms {
		if name := kinds[f.kind].name; strings.HasPrefix(suffix, name+"(") {
			return f, suffix[len(name):]
		}
	}
	// chordSuffix returns a kind, or a kind and tensions, and building
	// romanForms checked that every kind has a form.
	panic("chordSuffix returned " + suffix)
}

// ParseRoman reads a Roman numeral in the form that FormatRoman writes, such
// as "ii7", "♭VII", "V7(9)", "Vsus4" or "V7/V". Only the display forms are
// read: the accidentals ♭ and ♯, and the kinds and tensions as written. It
// fails with a *ParseError that says where reading stopped.
func ParseRoman(s string) (Position, error) {
	p, rest, err := readRoman(s)
	if err != nil {
		return Position{}, &ParseError{Input: s, Rest: rest, Err: err}
	}
	return p, nil
}

func readRoman(s string) (Position, string, error) {
	rest, alter := s, Natural
	switch {
	case strings.HasPrefix(rest, "♭"):
		rest, alter = rest[len("♭"):], Flat
	case strings.HasPrefix(rest, "♯"):
		rest, alter = rest[len("♯"):], Sharp
	}
	atNumeral := rest
	degree, c, afterNumeral, ok := readNumeral(atNumeral)
	if !ok {
		return Position{}, atNumeral, ErrNumeral
	}
	form, intervals, rest, err := readRomanKind(afterNumeral, c)
	if err != nil {
		return Position{}, rest, err
	}
	if form.c == byDegree && c != noThirdCase(degree, alter) {
		return Position{}, atNumeral, ErrRomanCase
	}
	// The intervals are those of the vocabulary on C, so this cannot fail.
	onC, _ := material.NewChord(material.NoteName{}, intervals...)
	// The degree and the alteration were read in range.
	p, _ := NewPosition(degree, alter, ShapeOf(onC))
	if rest == "" {
		return p, "", nil
	}
	return readTarget(p, rest)
}

// readRomanKind reads the kind written after a numeral in the case, and the
// tensions that follow it, up to the end or a "/". It returns the intervals of
// the members above the root and the rest; on failure, rest is where reading
// stopped.
func readRomanKind(s string, c letterCase) (romanForm, []material.Interval, string, error) {
	i := slices.IndexFunc(romanForms, func(f romanForm) bool {
		return (f.c == c || f.c == byDegree) && strings.HasPrefix(s, f.text)
	})
	if i < 0 {
		return romanForm{}, nil, s, ErrRomanKind
	}
	form, rest := romanForms[i], s[len(romanForms[i].text):]
	k := kinds[form.kind]
	intervals := slices.Clone(k.intervals)
	switch {
	case strings.HasPrefix(rest, "(") && !k.tensions:
		return romanForm{}, nil, rest, ErrNoTensions
	case strings.HasPrefix(rest, "("):
		tensions, after, err := readDisplayTensions(rest)
		if err != nil {
			return romanForm{}, nil, after, err
		}
		return form, append(intervals, tensions...), after, nil
	case rest != "" && !strings.HasPrefix(rest, "/"):
		return romanForm{}, nil, s, ErrRomanKind
	}
	return form, intervals, rest, nil
}

// readTarget reads "/" and the target that make up the rest of the numeral.
func readTarget(p Position, s string) (Position, string, error) {
	if !strings.HasPrefix(s, "/") {
		return Position{}, s, ErrTrailing
	}
	target, c, after, ok := readNumeral(s[1:])
	if !ok {
		return Position{}, s[1:], ErrTarget
	}
	if after != "" {
		return Position{}, after, ErrTrailing
	}
	p, err := p.WithTarget(target)
	if err != nil || c != majorKeyCase[target-1] {
		return Position{}, s[1:], ErrTarget
	}
	return p, "", nil
}

// readNumeral reads a numeral, I to VII or i to vii, at the start of s. The
// degrees are tried from 7 down, so that a numeral is tried before the shorter
// ones it starts with (VII before VI before V, III before II before I).
func readNumeral(s string) (int, letterCase, string, bool) {
	for _, c := range []letterCase{upper, lower} {
		for d := 7; d >= 1; d-- {
			if n := numeral(d, c); strings.HasPrefix(s, n) {
				return d, c, s[len(n):], true
			}
		}
	}
	return 0, upper, s, false
}

// readDisplayTensions reads "(t,t,...)" at the start of s, as readTensions
// does, but only the display names of the tensions (♭9, not b9).
func readDisplayTensions(s string) ([]material.Interval, string, error) {
	tensions, rest, err := readTensions(s)
	if err != nil {
		return nil, rest, err
	}
	item := s[1:]
	for _, t := range tensions {
		name := tensionNames[slices.IndexFunc(tensionNames, func(n tension) bool { return n.interval == t })].name
		if !strings.HasPrefix(item, name) {
			return nil, item, ErrTension
		}
		item = item[min(len(name)+1, len(item)):]
	}
	return tensions, rest, nil
}
