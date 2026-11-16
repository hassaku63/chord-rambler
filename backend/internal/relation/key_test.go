package relation_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// Alterations of note names, for the tables.
const (
	flat  = -1
	nat   = 0
	sharp = 1
)

// The tables hold only valid values, so the builders below panic on a mistake
// in a table.

func note(l material.Letter, alter int) material.NoteName {
	n, err := material.NewNoteName(l, alter)
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
func chord(root material.NoteName, intervals, bass string) material.Chord {
	c, err := material.NewChord(root, ivs(intervals)...)
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

// position builds the position of the shape (intervals and bass, as chord
// takes them) on the degree.
func position(degree int, alter material.Alteration, intervals, bass string) material.Position {
	p, err := material.NewPosition(degree, alter, material.ShapeOf(chord(material.NoteName{}, intervals, bass)))
	if err != nil {
		panic(err)
	}
	return p
}

func describe(p material.Position) string {
	bass, ok := p.Shape().Bass()
	return fmt.Sprintf("degree %d alteration %d %v bass %v %v", p.Degree(), p.Alteration(), p.Shape().Intervals(), bass, ok)
}

var cMajor = major(note(material.C, nat))

// The diatonic triads and seventh chords of C major: their positions, and
// that they are in the scale and in C Ionian. Sources: Hutchinson, Music
// Theory for the 21st-Century Classroom, 7.2 Diatonic Chords in Major
// (https://musictheory.pugetsound.edu/mt21c/DiatonicChordsInMajor.html):
// diatonic means within the key signature, and the triads follow the pattern
// M m m M M m °. The seventh chords add the seventh above the root from the
// scale; 8.3 Roman Numerals of Diatonic Seventh Chords
// (https://musictheory.pugetsound.edu/mt21c/RomanNumeralsOfDiatonicSeventhChords.html)
// gives the meaning of 7 (a minor seventh), M7 (a major seventh) and ø7 (a
// diminished triad with a minor seventh).
func TestDiatonicChordsOfCMajor(t *testing.T) {
	tests := []struct {
		name      string
		root      material.Letter
		degree    int
		intervals string
	}{
		{"I", material.C, 1, "M3 P5"},
		{"ii", material.D, 2, "m3 P5"},
		{"iii", material.E, 3, "m3 P5"},
		{"IV", material.F, 4, "M3 P5"},
		{"V", material.G, 5, "M3 P5"},
		{"vi", material.A, 6, "m3 P5"},
		{"vii°", material.B, 7, "m3 d5"},
		{"Imaj7", material.C, 1, "M3 P5 M7"},
		{"ii7", material.D, 2, "m3 P5 m7"},
		{"iii7", material.E, 3, "m3 P5 m7"},
		{"IVmaj7", material.F, 4, "M3 P5 M7"},
		{"V7", material.G, 5, "M3 P5 m7"},
		{"vi7", material.A, 6, "m3 P5 m7"},
		{"viiø7", material.B, 7, "m3 d5 m7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chord(note(tt.root, nat), tt.intervals, "")
			want := position(tt.degree, material.Natural, tt.intervals, "")
			if got, err := relation.PositionOf(c, cMajor); err != nil || got != want {
				t.Errorf("PositionOf = %s, %v; want %s", describe(got), err, describe(want))
			}
		})
	}
}

// Chords outside C major, and their positions (Hutchinson 19.1 Mode Mixture,
// https://musictheory.pugetsound.edu/mt21c/ModeMixtureSection.html, names iv
// and ♭VII among the chords borrowed from the parallel minor). The modes that
// hold their notes are in TestModesHolding.
func TestChordsOutsideCMajor(t *testing.T) {
	tests := []struct {
		name  string
		chord material.Chord
		want  material.Position
	}{
		{"iv (Fm)", chord(note(material.F, nat), "m3 P5", ""), position(4, material.Natural, "m3 P5", "")},
		{"♭VII (B♭)", chord(note(material.B, flat), "M3 P5", ""), position(7, material.Flat, "M3 P5", "")},
		{"♭VI (A♭)", chord(note(material.A, flat), "M3 P5", ""), position(6, material.Flat, "M3 P5", "")},
		{"C/B♭", chord(note(material.C, nat), "M3 P5", "m7"), position(1, material.Natural, "M3 P5", "m7")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := relation.PositionOf(tt.chord, cMajor); err != nil || got != tt.want {
				t.Errorf("PositionOf = %s, %v; want %s", describe(got), err, describe(tt.want))
			}
		})
	}
}

// The degree is counted by letters (in C, D♭ is the second degree and C♯ the
// first, #86), and a root two half steps from its degree has no position.
func TestPositionOfSpelling(t *testing.T) {
	triad := "M3 P5"
	tests := []struct {
		name  string
		chord material.Chord
		key   material.Scale
		want  material.Position
	}{
		{"D♭ in C is ♭II", chord(note(material.D, flat), triad, ""), cMajor, position(2, material.Flat, triad, "")},
		{"C♯ in C is ♯I", chord(note(material.C, sharp), triad, ""), cMajor, position(1, material.Sharp, triad, "")},
		{"F in D is ♭III", chord(note(material.F, nat), triad, ""), major(note(material.D, nat)), position(3, material.Flat, triad, "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := relation.PositionOf(tt.chord, tt.key); err != nil || got != tt.want {
				t.Errorf("PositionOf = %s, %v; want %s", describe(got), err, describe(tt.want))
			}
		})
	}
	aeolian, err := material.NewScale(note(material.C, nat), material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	rejects := []struct {
		name  string
		chord material.Chord
		key   material.Scale
		want  error
	}{
		{"E♭ in F♯ (the degree is E♯)", chord(note(material.E, flat), triad, ""), major(note(material.F, sharp)), relation.ErrNoPosition},
		{"C𝄪 in C", chord(note(material.C, 2), triad, ""), cMajor, relation.ErrNoPosition},
		{"a key in Aeolian", chord(note(material.C, nat), "m3 P5", ""), aeolian, relation.ErrNotMajor},
	}
	for _, tt := range rejects {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := relation.PositionOf(tt.chord, tt.key); !errors.Is(err, tt.want) {
				t.Errorf("PositionOf = %s, %v; want error %v", describe(got), err, tt.want)
			}
		})
	}
}

// The leading tone of C is B, a half step below the tonic (Hutchinson 3.3
// Scale Degree Names,
// https://musictheory.pugetsound.edu/mt21c/ScaleDegreeNames.html). Six half
// steps are an augmented fourth or a diminished fifth (Open Music Theory,
// Intervals and dyads, the table of chromatic and diatonic intervals: i6 is A4
// and d5, https://openmusictheory.github.io/intervals.html). The bass of C/B
// sounds, so it counts (#73).
func TestLeadingToneAndTritone(t *testing.T) {
	tests := []struct {
		name    string
		chord   material.Chord
		leading bool
		tritone bool
	}{
		{"G7 (B and B–F)", chord(note(material.G, nat), "M3 P5 m7", ""), true, true},
		{"Bdim (B and B–F)", chord(note(material.B, nat), "m3 d5", ""), true, true},
		{"Dm7 (neither)", chord(note(material.D, nat), "m3 P5 m7", ""), false, false},
		{"Em (B, no tritone)", chord(note(material.E, nat), "m3 P5", ""), true, false},
		{"D7 (F♯–C, no B)", chord(note(material.D, nat), "M3 P5 m7", ""), false, true},
		{"C/B (B in the bass)", chord(note(material.C, nat), "M3 P5", "M7"), true, false},
		// C♭ sounds like B but is not the leading tone of C: it is judged by spelling.
		{"A♭m (C♭, not B)", chord(note(material.A, flat), "m3 P5", ""), false, false},
		{"Fdim (C♭ is a member, F–C♭ a tritone)", chord(note(material.F, nat), "m3 d5", ""), false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relation.HasLeadingTone(tt.chord, cMajor); got != tt.leading {
				t.Errorf("HasLeadingTone = %v; want %v", got, tt.leading)
			}
			if got := relation.HasTritone(tt.chord); got != tt.tritone {
				t.Errorf("HasTritone = %v; want %v", got, tt.tritone)
			}
		})
	}
}

// V7 in the twelve major keys (docs/architecture/engine.md, "調"): the fifth
// degree of each major scale (Hutchinson 2.3 Major Key Signatures,
// https://musictheory.pugetsound.edu/mt21c/MajorKeySignatures.html) with a
// major triad and a minor seventh (8.3), and it holds the leading tone.
func TestDominantSeventhInMajorKeys(t *testing.T) {
	tests := []struct {
		tonic, dominant material.NoteName
	}{
		{note(material.C, nat), note(material.G, nat)},
		{note(material.G, nat), note(material.D, nat)},
		{note(material.D, nat), note(material.A, nat)},
		{note(material.A, nat), note(material.E, nat)},
		{note(material.E, nat), note(material.B, nat)},
		{note(material.B, nat), note(material.F, sharp)},
		{note(material.F, sharp), note(material.C, sharp)},
		{note(material.D, flat), note(material.A, flat)},
		{note(material.A, flat), note(material.E, flat)},
		{note(material.E, flat), note(material.B, flat)},
		{note(material.B, flat), note(material.F, nat)},
		{note(material.F, nat), note(material.C, nat)},
	}
	seventh := "M3 P5 m7"
	want := position(5, material.Natural, seventh, "")
	for _, tt := range tests {
		t.Run(tt.tonic.String(), func(t *testing.T) {
			c, key := chord(tt.dominant, seventh, ""), major(tt.tonic)
			if got, err := relation.PositionOf(c, key); err != nil || got != want {
				t.Errorf("PositionOf(%s7) = %s, %v; want V7", tt.dominant, describe(got), err)
			}
			if !relation.HasLeadingTone(c, key) {
				t.Errorf("HasLeadingTone(%s7) = false; want true", tt.dominant)
			}
			if !relation.HasTritone(c) {
				t.Errorf("HasTritone(%s7) = false; want true", tt.dominant)
			}
		})
	}
}

// Of the facts of one chord in a key, only the position is limited to major
// keys. The others do not look at the mode: the leading tone is the seventh
// degree of the major scale on the tonic, B in C, whatever the mode.
func TestKeyOtherThanIonian(t *testing.T) {
	cAeolian, err := material.NewScale(note(material.C, nat), material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	fm := chord(note(material.F, nat), "m3 P5", "")
	if _, err := relation.PositionOf(fm, cAeolian); !errors.Is(err, relation.ErrNotMajor) {
		t.Errorf("PositionOf in C Aeolian: err = %v; want ErrNotMajor", err)
	}
	g7 := chord(note(material.G, nat), "M3 P5 m7", "")
	if !relation.HasLeadingTone(g7, cAeolian) || relation.HasLeadingTone(fm, cAeolian) {
		t.Error("HasLeadingTone in C Aeolian: want true for G7 (B) and false for Fm")
	}
	if !relation.HasTritone(g7) || relation.HasTritone(fm) {
		t.Error("HasTritone: want true for G7 (B–F) and false for Fm")
	}
}

// The modes on a tonic that hold the notes by spelling. The sets are read off
// the modes on C in Hutchinson 31.9, Table 31.9.4
// (https://musictheory.pugetsound.edu/mt21c/JazzScales.html): C Ionian is
// C D E F G A B, Dorian C D E♭ F G A B♭, Phrygian C D♭ E♭ F G A♭ B♭, Lydian
// C D E F♯ G A B, Mixolydian C D E F G A B♭, Aeolian C D E♭ F G A♭ B♭ and
// Locrian C D♭ E♭ F G♭ A♭ B♭. A note is compared by its interval from the
// tonic, so A♯ is not B♭ (an augmented sixth is no sixth of a mode on C), and
// a tonic whose modes cannot be spelled is compared the same way, from its
// interval to each note. On G𝄫 the Ionian scale is G𝄫 A𝄫 B𝄫 C𝄫 D𝄫 E𝄫 F♭, and
// Lydian has C♭ (an augmented fourth) in place of C𝄫.
func TestModesHolding(t *testing.T) {
	all := []material.Mode{material.Ionian, material.Dorian, material.Phrygian, material.Lydian,
		material.Mixolydian, material.Aeolian, material.Locrian}
	tests := []struct {
		name  string
		notes []material.NoteName
		tonic material.NoteName
		want  []material.Mode
	}{
		{"no notes: every mode", nil, note(material.C, nat), all},
		{"the tonic only: every mode", []material.NoteName{note(material.C, nat)}, note(material.C, nat), all},
		{"C E G in C", []material.NoteName{note(material.C, nat), note(material.E, nat), note(material.G, nat)},
			note(material.C, nat),
			[]material.Mode{material.Ionian, material.Lydian, material.Mixolydian}},
		{"B♭ in C", []material.NoteName{note(material.B, flat)}, note(material.C, nat),
			[]material.Mode{material.Dorian, material.Phrygian, material.Mixolydian, material.Aeolian, material.Locrian}},
		{"A♯ in C: no mode (B♭ is spelled differently)", []material.NoteName{note(material.A, sharp)}, note(material.C, nat), nil},
		{"A♭ in C", []material.NoteName{note(material.A, flat)}, note(material.C, nat),
			[]material.Mode{material.Phrygian, material.Aeolian, material.Locrian}},
		{"B♭ and E in C", []material.NoteName{note(material.B, flat), note(material.E, nat)}, note(material.C, nat),
			[]material.Mode{material.Mixolydian}},
		{"F♯ in C", []material.NoteName{note(material.F, sharp)}, note(material.C, nat),
			[]material.Mode{material.Lydian}},
		{"G♭ in C", []material.NoteName{note(material.G, flat)}, note(material.C, nat),
			[]material.Mode{material.Locrian}},
		{"E, G♯, B on the tonic C𝄪: a diminished third, fifth and seventh, no mode",
			[]material.NoteName{note(material.E, nat), note(material.G, sharp), note(material.B, nat)},
			note(material.C, 2), nil},
		{"the major triad on the tonic G𝄫", []material.NoteName{note(material.G, -2), note(material.B, -2), note(material.D, -2)},
			note(material.G, -2),
			[]material.Mode{material.Ionian, material.Lydian, material.Mixolydian}},
		{"B𝄫 and F♭ on the tonic G𝄫: a major third and seventh", []material.NoteName{note(material.B, -2), note(material.F, -1)},
			note(material.G, -2),
			[]material.Mode{material.Ionian, material.Lydian}},
		{"C♭ on the tonic G𝄫: an augmented fourth", []material.NoteName{note(material.C, -1)},
			note(material.G, -2),
			[]material.Mode{material.Lydian}},
		{"B♭ on the tonic G𝄫: an augmented third, no mode", []material.NoteName{note(material.B, -1)},
			note(material.G, -2), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relation.ModesHolding(tt.notes, tt.tonic); !slices.Equal(got, tt.want) {
				t.Errorf("ModesHolding(%v, %s) = %v; want %v", tt.notes, tt.tonic, got, tt.want)
			}
		})
	}
}
