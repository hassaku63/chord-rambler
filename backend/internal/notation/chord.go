package notation

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrChordKind reports a chord kind outside the vocabulary or its aliases
// (docs/architecture/theory-conventions.md, 6), such as "-" or "9".
var ErrChordKind = errors.New("unknown chord kind")

// ErrUnclosedParen reports a "(" without its ")".
var ErrUnclosedParen = errors.New("parenthesis is not closed")

// ErrNoTensions reports tensions after a kind that takes none. Only 7, maj7
// and m7 take tensions.
var ErrNoTensions = errors.New("this chord kind takes no tensions")

// ErrTension reports an item in the parentheses that is not a tension.
var ErrTension = errors.New("not a tension: 9, ♭9, ♯9, 11, ♯11, 13 or ♭13")

// ErrTensionOrder reports tensions that repeat a number or are not in
// ascending order (9, then 11, then 13).
var ErrTensionOrder = errors.New("tensions must each appear once, in ascending order")

// ErrNoName reports a chord that has no name in the vocabulary, such as a
// fifth alone (C5) or a sixth and a ninth (C6/9).
var ErrNoName = errors.New("chord has no name in the vocabulary")

// kind is one chord kind of the vocabulary: the suffix written after the
// root, the other spellings read, and the intervals of the members above the
// root, in ascending number.
type kind struct {
	name      string
	aliases   []string
	intervals []material.Interval
	tensions  bool // whether tensions in parentheses may follow
}

func iv(number int, q material.Quality) material.Interval {
	i, err := material.NewInterval(number, q)
	if err != nil {
		panic(err) // the tables below hold only valid intervals
	}
	return i
}

var (
	maj3 = iv(3, material.Major)
	min3 = iv(3, material.Minor)
	p5   = iv(5, material.Perfect)
	min7 = iv(7, material.Minor)
)

// kinds is the vocabulary (docs/architecture/theory-conventions.md, 6).
var kinds = []kind{
	{"", nil, []material.Interval{maj3, p5}, false},
	{"m", []string{"min"}, []material.Interval{min3, p5}, false},
	{"dim", []string{"°"}, []material.Interval{min3, iv(5, material.Diminished)}, false},
	{"aug", []string{"+"}, []material.Interval{maj3, iv(5, material.Augmented)}, false},
	{"7", nil, []material.Interval{maj3, p5, min7}, true},
	{"maj7", []string{"M7", "Δ7"}, []material.Interval{maj3, p5, iv(7, material.Major)}, true},
	{"m7", []string{"min7"}, []material.Interval{min3, p5, min7}, true},
	{"m7(♭5)", []string{"m7♭5", "ø7", "min7(♭5)", "min7♭5"}, []material.Interval{min3, iv(5, material.Diminished), min7}, false},
	{"dim7", []string{"°7"}, []material.Interval{min3, iv(5, material.Diminished), iv(7, material.Diminished)}, false},
	{"sus2", nil, []material.Interval{iv(2, material.Major), p5}, false},
	{"sus4", nil, []material.Interval{iv(4, material.Perfect), p5}, false},
	{"7sus4", nil, []material.Interval{iv(4, material.Perfect), p5, min7}, false},
	{"add9", nil, []material.Interval{maj3, p5, iv(9, material.Major)}, false},
	{"6", nil, []material.Interval{maj3, p5, iv(6, material.Major)}, false},
	{"m6", []string{"min6"}, []material.Interval{min3, p5, iv(6, material.Major)}, false},
	{"m(add9)", []string{"min(add9)"}, []material.Interval{min3, p5, iv(9, material.Major)}, false},
}

// tension is a ninth, eleventh or thirteenth with its written name.
type tension struct {
	name     string
	interval material.Interval
}

// tensionNames holds the tensions with their written names.
var tensionNames = []tension{
	{"9", iv(9, material.Major)}, {"♭9", iv(9, material.Minor)}, {"♯9", iv(9, material.Augmented)},
	{"11", iv(11, material.Perfect)}, {"♯11", iv(11, material.Augmented)},
	{"13", iv(13, material.Major)}, {"♭13", iv(13, material.Minor)},
}

type spelling struct {
	text  string
	index int
}

// kindSpellings and tensionSpellings hold every spelling read, each with the
// index of its kind or tension. ASCII # and b are read for ♯ and ♭. Kinds are
// longest first, so that "m7" is tried before "m" and "" comes last.
var kindSpellings, tensionSpellings = func() ([]spelling, []spelling) {
	var ks, ts []spelling
	for i, k := range kinds {
		for _, text := range append([]string{k.name}, k.aliases...) {
			ks = append(ks, withASCII(text, i)...)
		}
	}
	slices.SortStableFunc(ks, func(a, b spelling) int { return len(b.text) - len(a.text) })
	for i, t := range tensionNames {
		ts = append(ts, withASCII(t.name, i)...)
	}
	return ks, ts
}()

var toASCII = strings.NewReplacer("♯", "#", "♭", "b")

func withASCII(text string, index int) []spelling {
	out := []spelling{{text, index}}
	if a := toASCII.Replace(text); a != text {
		out = append(out, spelling{a, index})
	}
	return out
}

// ParseChord reads a chord name: a root (as ParseNoteName reads it), a kind,
// tensions in parentheses after 7, maj7 and m7, and a bass after "/", such as
// "C", "F♯m7(♭5)", "Bbmaj7(9,13)" or "C/E". The kinds and their aliases are
// those of docs/architecture/theory-conventions.md, 6. It fails with a
// *ParseError that says where reading stopped.
func ParseChord(s string) (material.Chord, error) {
	c, rest, err := readChord(s)
	if err != nil {
		return material.Chord{}, &ParseError{Input: s, Rest: rest, Err: err}
	}
	return c, nil
}

func readChord(s string) (material.Chord, string, error) {
	root, rest, err := readNoteName(s)
	if err != nil {
		return material.Chord{}, rest, err
	}
	afterRoot := rest
	var k kind
	for _, sp := range kindSpellings {
		if strings.HasPrefix(rest, sp.text) {
			k, rest = kinds[sp.index], rest[len(sp.text):]
			break
		}
	}
	intervals := slices.Clone(k.intervals)
	if strings.HasPrefix(rest, "(") {
		if !k.tensions {
			return material.Chord{}, rest, ErrNoTensions
		}
		var tensions []material.Interval
		if tensions, rest, err = readTensions(rest); err != nil {
			return material.Chord{}, rest, err
		}
		intervals = append(intervals, tensions...)
	} else if rest != "" && !strings.HasPrefix(rest, "/") {
		return material.Chord{}, afterRoot, ErrChordKind
	}
	c, err := material.NewChord(root, intervals...)
	if err != nil {
		return material.Chord{}, s, err
	}
	if rest == "" {
		return c, "", nil
	}
	if !strings.HasPrefix(rest, "/") {
		return material.Chord{}, rest, ErrTrailing
	}
	return readBass(c, rest[1:])
}

// readTensions reads "(t,t,...)" at the start of s.
func readTensions(s string) ([]material.Interval, string, error) {
	end := strings.Index(s, ")")
	if end < 0 {
		return nil, s, ErrUnclosedParen
	}
	var tensions []material.Interval
	item := s[1:]
	for _, text := range strings.Split(s[1:end], ",") {
		i := slices.IndexFunc(tensionSpellings, func(sp spelling) bool { return sp.text == text })
		if i < 0 {
			return nil, item, ErrTension
		}
		t := tensionNames[tensionSpellings[i].index].interval
		if len(tensions) > 0 && t.Number() <= tensions[len(tensions)-1].Number() {
			return nil, item, ErrTensionOrder
		}
		tensions = append(tensions, t)
		item = item[min(len(text)+1, len(item)):]
	}
	return tensions, s[end+1:], nil
}

// readBass reads the bass note name that makes up the rest of the name.
func readBass(c material.Chord, s string) (material.Chord, string, error) {
	bass, rest, err := readNoteName(s)
	if err != nil {
		return material.Chord{}, rest, err
	}
	if rest != "" {
		return material.Chord{}, rest, ErrTrailing
	}
	between, err := material.IntervalBetween(c.Root(), bass)
	if err != nil {
		return material.Chord{}, s, err
	}
	if c, err = c.WithBass(between); err != nil {
		return material.Chord{}, s, err
	}
	return c, "", nil
}

// FormatChord writes the chord name in the display form, keeping the spelling
// of the root and the bass (B𝄫m stays B𝄫m), so that ParseChord reads it back
// to the same chord. Chord names in a key, shown with Roman numerals, are
// written with it (♭VII on D♭ is C♭), so that the letter follows the degree. It
// fails with ErrNoName for a chord outside the vocabulary.
func FormatChord(c material.Chord) (string, error) {
	return formatChord(c, false)
}

// FormatChordForDisplay writes the chord name as FormatChord does, but with the
// root respelled by DisplaySpelling (B♭♭m is Am). When the root is respelled,
// the bass is written from the new root at the same interval, so the bass keeps
// its interval to the root (B♭♭/D♭ is A/C♯); otherwise the bass is respelled on
// its own. The name may read back to a different spelling of the chord. It is
// only for a chord shown without a key; in a key, use FormatChord
// (docs/architecture/theory-conventions.md, 6).
func FormatChordForDisplay(c material.Chord) (string, error) {
	return formatChord(c, true)
}

func formatChord(c material.Chord, display bool) (string, error) {
	suffix, err := chordSuffix(c.Intervals()[1:])
	if err != nil {
		return "", fmt.Errorf("chord on %s with %v: %w", c.Root(), c.Intervals()[1:], err)
	}
	root := c.Root()
	if display {
		root = DisplaySpelling(root)
	}
	name := FormatNoteName(root) + suffix
	if bass, ok := c.Bass(); ok {
		n, err := root.Transpose(bass)
		if err != nil {
			return "", err
		}
		if display && root == c.Root() {
			n = DisplaySpelling(n)
		}
		name += "/" + FormatNoteName(n)
	}
	return name, nil
}

// chordSuffix returns the kind and the tensions for the intervals above the
// root.
func chordSuffix(intervals []material.Interval) (string, error) {
	for _, k := range kinds {
		if slices.Equal(k.intervals, intervals) {
			return k.name, nil
		}
	}
	split := slices.IndexFunc(intervals, func(i material.Interval) bool { return i.Number() > 7 })
	if split < 0 {
		return "", ErrNoName
	}
	base, tensions := intervals[:split], intervals[split:]
	k := slices.IndexFunc(kinds, func(k kind) bool { return k.tensions && slices.Equal(k.intervals, base) })
	if k < 0 {
		return "", ErrNoName
	}
	names := make([]string, 0, len(tensions))
	for _, t := range tensions {
		i := slices.IndexFunc(tensionNames, func(n tension) bool { return n.interval == t })
		if i < 0 {
			return "", ErrNoName
		}
		names = append(names, tensionNames[i].name)
	}
	return kinds[k].name + "(" + strings.Join(names, ",") + ")", nil
}
