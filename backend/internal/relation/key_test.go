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
			if !relation.InScale(c, cMajor) {
				t.Errorf("InScale = false; want true")
			}
			if got := relation.ParallelModes(c, cMajor); !slices.Contains(got, material.Ionian) {
				t.Errorf("ParallelModes = %v; want Ionian among them", got)
			}
		})
	}
}

// Chords outside C major, and the modes on C that contain what sounds. The
// sets are read off the modes on C in Hutchinson 31.9, Table 31.9.4
// (https://musictheory.pugetsound.edu/mt21c/JazzScales.html): A♭ is in
// Phrygian, Aeolian and Locrian only; B♭ with D natural in Dorian, Mixolydian
// and Aeolian; B♭ with E natural in Mixolydian only. Hutchinson 19.1 Mode
// Mixture (https://musictheory.pugetsound.edu/mt21c/ModeMixtureSection.html)
// names iv and ♭VII among the chords borrowed from the parallel minor. C/B♭ is
// the fact of #73: by its members it is in Ionian, Lydian and Mixolydian, by
// what sounds in Mixolydian only.
func TestChordsOutsideCMajor(t *testing.T) {
	tests := []struct {
		name  string
		chord material.Chord
		want  material.Position
		modes []material.Mode
	}{
		{"iv (Fm)", chord(note(material.F, nat), "m3 P5", ""), position(4, material.Natural, "m3 P5", ""),
			[]material.Mode{material.Phrygian, material.Aeolian, material.Locrian}},
		{"♭VII (B♭)", chord(note(material.B, flat), "M3 P5", ""), position(7, material.Flat, "M3 P5", ""),
			[]material.Mode{material.Dorian, material.Mixolydian, material.Aeolian}},
		{"♭VI (A♭)", chord(note(material.A, flat), "M3 P5", ""), position(6, material.Flat, "M3 P5", ""),
			[]material.Mode{material.Phrygian, material.Aeolian, material.Locrian}},
		{"C/B♭", chord(note(material.C, nat), "M3 P5", "m7"), position(1, material.Natural, "M3 P5", "m7"),
			[]material.Mode{material.Mixolydian}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := relation.PositionOf(tt.chord, cMajor); err != nil || got != tt.want {
				t.Errorf("PositionOf = %s, %v; want %s", describe(got), err, describe(tt.want))
			}
			if relation.InScale(tt.chord, cMajor) {
				t.Errorf("InScale = true; want false")
			}
			if got := relation.ParallelModes(tt.chord, cMajor); !slices.Equal(got, tt.modes) {
				t.Errorf("ParallelModes = %v; want %v", got, tt.modes)
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
			if !relation.InScale(c, key) {
				t.Errorf("InScale(%s7, %s major) = false; want true", tt.dominant, tt.tonic)
			}
		})
	}
}

// Of the facts of one chord in a key, only the position is limited to major
// keys. The other four do not look at the mode, so in C Aeolian
// (C D E♭ F G A♭ B♭) Fm is in the scale and its parallel modes are those on C
// whose scale holds F, A♭ and C: Phrygian, Aeolian and Locrian. The leading tone is the seventh degree of the major
// scale on the tonic, B, whatever the mode.
func TestKeyOtherThanIonian(t *testing.T) {
	cAeolian, err := material.NewScale(note(material.C, nat), material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	fm := chord(note(material.F, nat), "m3 P5", "")
	if _, err := relation.PositionOf(fm, cAeolian); !errors.Is(err, relation.ErrNotMajor) {
		t.Errorf("PositionOf in C Aeolian: err = %v; want ErrNotMajor", err)
	}
	if !relation.InScale(fm, cAeolian) {
		t.Error("InScale(Fm, C Aeolian) = false; want true")
	}
	if relation.InScale(chord(note(material.F, nat), "M3 P5", ""), cAeolian) {
		t.Error("InScale(F, C Aeolian) = true; want false (A is not in the scale)")
	}
	want := []material.Mode{material.Phrygian, material.Aeolian, material.Locrian}
	if got := relation.ParallelModes(fm, cAeolian); !slices.Equal(got, want) {
		t.Errorf("ParallelModes(Fm, C Aeolian) = %v; want %v", got, want)
	}
	g7 := chord(note(material.G, nat), "M3 P5 m7", "")
	if !relation.HasLeadingTone(g7, cAeolian) || relation.HasLeadingTone(fm, cAeolian) {
		t.Error("HasLeadingTone in C Aeolian: want true for G7 (B) and false for Fm")
	}
	if !relation.HasTritone(g7) || relation.HasTritone(fm) {
		t.Error("HasTritone: want true for G7 (B–F) and false for Fm")
	}
}

// A tonic that cannot spell its modes within a double sharp is still compared
// by pitch class. On C𝄪 (pitch class 2) the E major triad (E, G♯, B: 4, 8, 11)
// is 2, 6, 9 above the tonic, and of the seven modes only Lydian (0 2 4 6 7 9
// 11) holds all three: Ionian lacks 6, the others lack 6 or 9.
func TestParallelModesOfUnspellableMode(t *testing.T) {
	key := major(note(material.C, 2))
	e := chord(note(material.E, nat), "M3 P5", "")
	want := []material.Mode{material.Lydian}
	if got := relation.ParallelModes(e, key); !slices.Equal(got, want) {
		t.Errorf("ParallelModes(E, C𝄪 major) = %v; want %v", got, want)
	}
}
