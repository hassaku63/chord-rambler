package notation

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

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
func noThirdCase(degree int, alter material.Alteration) letterCase {
	if alter != material.Natural {
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
func FormatRoman(p material.Position) (string, error) {
	shape := p.Shape()
	if _, ok := shape.Bass(); ok {
		return "", ErrRomanBass
	}
	suffix, err := chordSuffix(shape.Intervals()[1:])
	if err != nil {
		return "", fmt.Errorf("shape %v: %w", shape.Intervals()[1:], err)
	}
	form, tensions := romanFormOf(suffix)
	c := form.c
	if c == byDegree {
		c = noThirdCase(p.Degree(), p.Alteration())
	}
	s := [...]string{"♭", "", "♯"}[p.Alteration()+1] + numeral(p.Degree(), c) + form.text + tensions
	if target, ok := p.Target(); ok {
		s += "/" + numeral(target, majorKeyCase[target-1])
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
func ParseRoman(s string) (material.Position, error) {
	p, rest, err := readRoman(s)
	if err != nil {
		return material.Position{}, &ParseError{Input: s, Rest: rest, Err: err}
	}
	return p, nil
}

func readRoman(s string) (material.Position, string, error) {
	rest, alter := s, material.Natural
	switch {
	case strings.HasPrefix(rest, "♭"):
		rest, alter = rest[len("♭"):], material.Flat
	case strings.HasPrefix(rest, "♯"):
		rest, alter = rest[len("♯"):], material.Sharp
	}
	atNumeral := rest
	degree, c, afterNumeral, ok := readNumeral(atNumeral)
	if !ok {
		return material.Position{}, atNumeral, ErrNumeral
	}
	form, intervals, rest, err := readRomanKind(afterNumeral, c)
	if err != nil {
		return material.Position{}, rest, err
	}
	if form.c == byDegree && c != noThirdCase(degree, alter) {
		return material.Position{}, atNumeral, ErrRomanCase
	}
	// The intervals are those of the vocabulary on C, so this cannot fail.
	onC, _ := material.NewChord(material.NoteName{}, intervals...)
	// The degree and the alteration were read in range.
	p, _ := material.NewPosition(degree, alter, material.ShapeOf(onC))
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
// The target is a degree 2 to 6 without an accidental, written in the case of
// its triad in the major key (ii, iii, IV, V, vi); anything else is
// material.ErrTarget. Text after the one target, as in V7/V/V, is ErrTrailing.
func readTarget(p material.Position, s string) (material.Position, string, error) {
	if !strings.HasPrefix(s, "/") {
		return material.Position{}, s, ErrTrailing
	}
	target, c, after, ok := readNumeral(s[1:])
	if !ok {
		return material.Position{}, s[1:], material.ErrTarget
	}
	if after != "" {
		return material.Position{}, after, ErrTrailing
	}
	p, err := p.WithTarget(target)
	if err != nil || c != majorKeyCase[target-1] {
		return material.Position{}, s[1:], material.ErrTarget
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
