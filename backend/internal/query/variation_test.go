package query_test

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// The tables hold only valid values, so the builders below panic on a mistake
// in a table.

func note(s string) material.NoteName {
	n, err := notation.ParseNoteName(s)
	if err != nil {
		panic(err)
	}
	return n
}

// ivs parses intervals such as "M3 P5 m7".
func ivs(s string) []material.Interval {
	var out []material.Interval
	for _, f := range strings.Fields(s) {
		q, ok := map[byte]material.Quality{
			'P': material.Perfect, 'M': material.Major, 'm': material.Minor,
			'A': material.Augmented, 'd': material.Diminished,
		}[f[0]]
		number, err := strconv.Atoi(f[1:])
		if !ok || err != nil {
			panic("bad interval in a test table: " + f)
		}
		iv, err := material.NewInterval(number, q)
		if err != nil {
			panic(err)
		}
		out = append(out, iv)
	}
	return out
}

// chord builds the chord on the root with the intervals and, if bass is not
// empty, the bass that interval above the root.
func chord(root, intervals, bass string) material.Chord {
	c, err := material.NewChord(note(root), ivs(intervals)...)
	if err != nil {
		panic(err)
	}
	if bass != "" {
		if c, err = c.WithBass(ivs(bass)[0]); err != nil {
			panic(err)
		}
	}
	return c
}

func major(tonic material.NoteName) material.Scale {
	s, err := material.NewScale(tonic, material.Ionian)
	if err != nil {
		panic(err)
	}
	return s
}

// names writes the notes as "G B D F♯", for the tables and the messages.
func names(notes []material.NoteName) string {
	out := make([]string, 0, len(notes))
	for _, n := range notes {
		out = append(out, notation.FormatNoteName(n))
	}
	return strings.Join(out, " ")
}

// read is the main reading expected, without the key: every test checks that
// the key is the one read in.
type read struct {
	kind     reading.Kind
	function reading.Function
	target   int
	modes    []material.Mode
}

func dia(f reading.Function) read { return read{kind: reading.Diatonic, function: f} }
func sec(target int) read         { return read{kind: reading.SecondaryDominant, target: target} }
func bor(modes ...material.Mode) read {
	return read{kind: reading.Borrowed, modes: modes}
}

var facts = read{kind: reading.FactsOnly}

// variation is one expected variation: its members (on the root of the chord
// asked about, no bass), whether it is the chord asked about, its main reading
// and its notes from outside the key (spellings not in the scale).
type variation struct {
	members  string
	input    bool
	read     read
	outOfKey string
}

// checkVariations checks the list in order. Every variation of these lists has
// no tension caution: the only tension in the lists is a major ninth, which is
// never a half step above the root, the third, the fifth or the seventh of
// these shapes.
func checkVariations(t *testing.T, c material.Chord, key material.Scale, want []variation) {
	t.Helper()
	got, err := query.Variations(c, key)
	if err != nil {
		t.Fatalf("Variations: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d variations; want %d", len(got), len(want))
	}
	for i, w := range want {
		g := got[i]
		if g.Chord.Root() != c.Root() {
			t.Errorf("%d (%s): root %s; want %s", i, w.members, g.Chord.Root(), c.Root())
		}
		if _, ok := g.Chord.Bass(); ok {
			t.Errorf("%d (%s): has a bass", i, w.members)
		}
		if s := names(g.Members.Notes); s != w.members {
			t.Errorf("%d: members %s; want %s", i, s, w.members)
		}
		if g.IsInput != w.input {
			t.Errorf("%d (%s): IsInput = %v; want %v", i, w.members, g.IsInput, w.input)
		}
		m := g.Reading.Main
		if m.Key != key {
			t.Errorf("%d (%s): read in %s; want %s", i, w.members, m.Key.Tonic(), key.Tonic())
		}
		if m.Kind != w.read.kind || m.Function != w.read.function || m.Target != w.read.target || !slices.Equal(m.Modes, w.read.modes) {
			t.Errorf("%d (%s): Kind, Function, Target, Modes = %s, %s, %d, %v; want %s, %s, %d, %v", i, w.members,
				m.Kind, m.Function, m.Target, m.Modes, w.read.kind, w.read.function, w.read.target, w.read.modes)
		}
		if len(g.Cautions.Cautions) != 0 {
			t.Errorf("%d (%s): cautions %v; want none", i, w.members, g.Cautions.Cautions)
		}
		if s := names(g.OutOfKey.Notes); s != w.outOfKey {
			t.Errorf("%d (%s): notes from outside the key %q; want %q", i, w.members, s, w.outOfKey)
		}
	}
}

// The readings are worked out by the rules of docs/architecture/domain-model.md
// (4, 和音 1 つの読み) from the modes on C:
//
//	Ionian     C D E F G A B
//	Dorian     C D E♭ F G A B♭
//	Phrygian   C D♭ E♭ F G A♭ B♭
//	Lydian     C D E F♯ G A B
//	Mixolydian C D E F G A B♭
//	Aeolian    C D E♭ F G A♭ B♭
//	Locrian    C D♭ E♭ F G♭ A♭ B♭
//
// The kind looks at the second to the seventh members, not at the ninth, and
// a sixth or a sus member keeps a chord from the shape of a secondary dominant
// (#87, #102). The lists are the owner's (#115).
func TestVariationsInCMajor(t *testing.T) {
	c := major(note("C"))
	lydian := bor(material.Lydian)
	gList := func(input int) []variation {
		vs := []variation{
			{"G B D", false, dia(reading.Dominant), ""},
			{"G B D F♯", false, lydian, "F♯"},
			{"G B D F", false, dia(reading.Dominant), ""},
			{"G B D E", false, dia(reading.Dominant), ""},
			{"G B D A", false, dia(reading.Dominant), ""},
			{"G A D", false, dia(reading.Dominant), ""},
			{"G C D", false, dia(reading.Dominant), ""},
			{"G C D F", false, dia(reading.Dominant), ""},
			{"G B D F A", false, dia(reading.Dominant), ""},
		}
		vs[input].input = true
		return vs
	}
	fromMinor := bor(material.Phrygian, material.Aeolian, material.Locrian)
	tests := []struct {
		name  string
		chord material.Chord
		want  []variation
	}{
		{"G", chord("G", "M3 P5", ""), gList(0)},
		// G7 has the same list, and the mark moves to G7. Gmaj7 has F♯.
		{"G7", chord("G", "M3 P5 m7", ""), gList(2)},
		{"Am", chord("A", "m3 P5", ""), []variation{
			{"A C E", true, dia(reading.Tonic), ""},
			{"A C E G", false, dia(reading.Tonic), ""},
			{"A C E F♯", false, lydian, "F♯"}, // the sixth is judged (#102)
			{"A C E B", false, dia(reading.Tonic), ""},
			{"A C E G B", false, dia(reading.Tonic), ""},
			{"A B E", false, dia(reading.Tonic), ""},
			{"A D E", false, dia(reading.Tonic), ""},
		}},
		{"Bdim", chord("B", "m3 d5", ""), []variation{
			{"B D F", true, dia(reading.Dominant), ""},
			{"B D F A", false, dia(reading.Dominant), ""},
			// A♭ and B together are in no church mode on C (#102, #105).
			{"B D F A♭", false, facts, "A♭"},
		}},
		// Em(add9) and Em9 have F♯ and are the diatonic tonic all the same:
		// the kind does not look at the ninth. The F♯ of Esus2 is the second,
		// which is judged, and so is the C♯ of Em6.
		{"Em", chord("E", "m3 P5", ""), []variation{
			{"E G B", true, dia(reading.Tonic), ""},
			{"E G B D", false, dia(reading.Tonic), ""},
			{"E G B C♯", false, facts, "C♯"},
			{"E G B F♯", false, dia(reading.Tonic), "F♯"},
			{"E G B D F♯", false, dia(reading.Tonic), "F♯"},
			{"E F♯ B", false, lydian, "F♯"},
			{"E A B", false, dia(reading.Tonic), ""},
		}},
		// Fm is borrowed already; every note it has outside the key counts.
		{"Fm", chord("F", "m3 P5", ""), []variation{
			{"F A♭ C", true, fromMinor, "A♭"},
			{"F A♭ C E♭", false, fromMinor, "A♭ E♭"},
			{"F A♭ C D", false, bor(material.Aeolian), "A♭"},
			{"F A♭ C G", false, fromMinor, "A♭"},
			{"F A♭ C E♭ G", false, fromMinor, "A♭ E♭"},
			{"F G C", false, dia(reading.Subdominant), ""},
			// The sus4 adds B♭ from outside the key, and is borrowed.
			{"F B♭ C", false, bor(material.Dorian, material.Phrygian, material.Mixolydian, material.Aeolian, material.Locrian), "B♭"},
		}},
		// D7 is V7/V. The triad, add9 and 9 keep the shape; maj7, 6 and the
		// sus chords do not.
		{"D7", chord("D", "M3 P5 m7", ""), []variation{
			{"D F♯ A", false, sec(5), "F♯"},
			{"D F♯ A C♯", false, facts, "F♯ C♯"},
			{"D F♯ A C", true, sec(5), "F♯"},
			{"D F♯ A B", false, lydian, "F♯"},
			{"D F♯ A E", false, sec(5), "F♯"},
			{"D E A", false, dia(reading.Subdominant), ""},
			{"D G A", false, dia(reading.Subdominant), ""},
			{"D G A C", false, dia(reading.Subdominant), ""},
			{"D F♯ A C E", false, sec(5), "F♯"},
		}},
		// The bass is left out: C/E has the list of C, and none of the
		// variations, having no bass, is C/E itself.
		{"C/E", chord("C", "M3 P5", "M3"), []variation{
			{"C E G", false, dia(reading.Tonic), ""},
			{"C E G B", false, dia(reading.Tonic), ""},
			{"C E G B♭", false, sec(4), "B♭"},
			{"C E G A", false, dia(reading.Tonic), ""},
			{"C E G D", false, dia(reading.Tonic), ""},
			{"C D G", false, dia(reading.Tonic), ""},
			{"C F G", false, dia(reading.Tonic), ""},
			{"C F G B♭", false, bor(material.Dorian, material.Phrygian, material.Mixolydian, material.Aeolian), "B♭"},
			{"C E G B♭ D", false, sec(4), "B♭"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { checkVariations(t, tt.chord, c, tt.want) })
	}
}

// A chord without a third (sus, a power chord) and an augmented triad have no
// quality to choose the list by.
func TestVariationsOfAChordWithoutQualityFail(t *testing.T) {
	tests := []struct {
		name  string
		chord material.Chord
	}{
		{"Gsus4", chord("G", "P4 P5", "")},
		{"Csus2", chord("C", "M2 P5", "")},
		{"C5", chord("C", "P5", "")},
		{"Caug", chord("C", "M3 A5", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := query.Variations(tt.chord, major(note("C")))
			if !errors.Is(err, query.ErrNoTriadQuality) {
				t.Errorf("err = %v; want ErrNoTriadQuality", err)
			}
		})
	}
}

func TestVariationsInAKeyOtherThanMajorFail(t *testing.T) {
	dorian, err := material.NewScale(note("D"), material.Dorian)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := query.Variations(chord("D", "m3 P5", ""), dorian); !errors.Is(err, relation.ErrNotMajor) {
		t.Errorf("err = %v; want relation.ErrNotMajor", err)
	}
}

// dim7 on C♭ would need B triple flat.
func TestVariationsBeyondADoubleFlatFail(t *testing.T) {
	_, err := query.Variations(chord("C♭", "m3 d5", ""), major(note("C")))
	if !errors.Is(err, material.ErrAlterOutOfRange) {
		t.Errorf("err = %v; want material.ErrAlterOutOfRange", err)
	}
}

// Each label carries the kind of knowledge it rests on.
func TestLabelBases(t *testing.T) {
	tests := []struct {
		name string
		got  query.Basis
		want query.Basis
	}{
		{"members", query.MembersLabel{}.Basis(), query.Definition},
		{"reading", query.ReadingLabel{}.Basis(), query.School},
		{"cautions", query.CautionsLabel{}.Basis(), query.School},
		{"out of key", query.OutOfKeyLabel{}.Basis(), query.Computation},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: %s; want %s", tt.name, tt.got, tt.want)
		}
	}
}
