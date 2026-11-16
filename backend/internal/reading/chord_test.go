package reading_test

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
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

// Shapes, as intervals above the root.
const (
	maj   = "M3 P5"
	min   = "m3 P5"
	dim   = "m3 d5"
	dom7  = "M3 P5 m7"
	maj7  = "M3 P5 M7"
	min7  = "m3 P5 m7"
	half7 = "m3 d5 m7"
)

// want is the main reading expected, without the key: every test checks that
// the key is the one read in.
type want struct {
	kind     reading.Kind
	function reading.Function
	target   int
	modes    []material.Mode
	minor    bool // FromParallelMinor
}

type row struct {
	name  string
	chord material.Chord
	want  want
}

func check(t *testing.T, key material.Scale, tests []row) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reading.ReadChord(tt.chord, key)
			if err != nil {
				t.Fatalf("ReadChord: %v", err)
			}
			m := got.Main
			if m.Key != key {
				t.Errorf("Key = %s %s; want %s %s", m.Key.Tonic(), m.Key.Mode(), key.Tonic(), key.Mode())
			}
			if m.Kind != tt.want.kind || m.Function != tt.want.function || m.Target != tt.want.target {
				t.Errorf("Kind, Function, Target = %s, %s, %d; want %s, %s, %d",
					m.Kind, m.Function, m.Target, tt.want.kind, tt.want.function, tt.want.target)
			}
			if !slices.Equal(m.Modes, tt.want.modes) {
				t.Errorf("Modes = %v; want %v", m.Modes, tt.want.modes)
			}
			if m.FromParallelMinor() != tt.want.minor {
				t.Errorf("FromParallelMinor = %v; want %v", m.FromParallelMinor(), tt.want.minor)
			}
			if len(got.Others) != 0 {
				t.Errorf("Others = %v; want none (#101, B1)", got.Others)
			}
		})
	}
}

var (
	cNat  = note(material.C, nat)
	dNat  = note(material.D, nat)
	eNat  = note(material.E, nat)
	fNat  = note(material.F, nat)
	gNat  = note(material.G, nat)
	aNat  = note(material.A, nat)
	bNat  = note(material.B, nat)
	bFlat = note(material.B, flat)
	aFlat = note(material.A, flat)
	eFlat = note(material.E, flat)
	dFlat = note(material.D, flat)
)

// The diatonic triads and seventh chords of C major (Hutchinson 7.2,
// https://musictheory.pugetsound.edu/mt21c/DiatonicChordsInMajor.html, and 8.3,
// https://musictheory.pugetsound.edu/mt21c/RomanNumeralsOfDiatonicSeventhChords.html) and
// their functions: I, iii and vi are T, ii and IV S, V and vii° D
// (docs/architecture/theory-conventions.md, 1; Yamaha Music Pal,
// https://jp.yamaha.com/services/music-pal/study/chord/cadence/index.html).
// The function follows the degree whatever the shape: Gsus4 and C/D are
// diatonic too (#101, D1) and read as D and T (#102).
//
// The kind is judged by the notes that make the chord and a bass a minor
// seventh above the root, by spelling (the owner's call on #102): a tension
// or another bass outside the scale does not change it. G7(♭9) is V, C(♯11)
// and C/F♯ are I, and F/D♯ is IV, its bass being an augmented sixth.
func TestDiatonicInCMajor(t *testing.T) {
	d := func(f reading.Function) want { return want{kind: reading.Diatonic, function: f} }
	check(t, major(cNat), []row{
		{"C", chord(cNat, maj, ""), d(reading.Tonic)},
		{"Dm", chord(dNat, min, ""), d(reading.Subdominant)},
		{"Em", chord(eNat, min, ""), d(reading.Tonic)},
		{"F", chord(fNat, maj, ""), d(reading.Subdominant)},
		{"G", chord(gNat, maj, ""), d(reading.Dominant)},
		{"Am", chord(aNat, min, ""), d(reading.Tonic)},
		{"Bdim", chord(bNat, dim, ""), d(reading.Dominant)},
		{"Cmaj7", chord(cNat, maj7, ""), d(reading.Tonic)},
		{"Dm7", chord(dNat, min7, ""), d(reading.Subdominant)},
		{"Em7", chord(eNat, min7, ""), d(reading.Tonic)},
		{"Fmaj7", chord(fNat, maj7, ""), d(reading.Subdominant)},
		{"G7", chord(gNat, dom7, ""), d(reading.Dominant)},
		{"Am7", chord(aNat, min7, ""), d(reading.Tonic)},
		{"Bm7(♭5)", chord(bNat, half7, ""), d(reading.Dominant)},
		{"Gsus4", chord(gNat, "P4 P5", ""), d(reading.Dominant)},
		{"C/D", chord(cNat, maj, "M2"), d(reading.Tonic)},
		{"G7(♭9)", chord(gNat, "M3 P5 m7 m9", ""), d(reading.Dominant)},
		{"C(♯11)", chord(cNat, "M3 P5 A11", ""), d(reading.Tonic)},
		{"C/F♯", chord(cNat, maj, "A4"), d(reading.Tonic)},
		{"F/D♯", chord(fNat, maj, "A6"), d(reading.Subdominant)},
	})
}

// Borrowed chords carry no function (#101, A1). The modes are those on C whose
// notes hold, by spelling, the notes that make the chord and a bass a minor
// seventh above the root (the owner's call on #102), worked out from the
// scales:
//
//	Dorian     C D E♭ F G A B♭
//	Phrygian   C D♭ E♭ F G A♭ B♭
//	Lydian     C D E F♯ G A B
//	Mixolydian C D E F G A B♭
//	Aeolian    C D E♭ F G A♭ B♭
//	Locrian    C D♭ E♭ F G♭ A♭ B♭
//
// Chords borrowed from the parallel minor: Hutchinson 19.1
// (https://musictheory.pugetsound.edu/mt21c/ModeMixtureSection.html) names
// i, iv and ♭VII7 and the lowered degrees ♭6, ♭3 and ♭7, and 19.2
// (https://musictheory.pugetsound.edu/mt21c/HarmonizationOfBorrowedScaleDegrees.html)
// takes each lowered degree as the root of a borrowed chord (♭VI, ♭III, ♭VII).
// iv and i are borrowed although their roots are not altered (#102).
// B♭maj7 (A, not A♭) and D♭ (D♭, not D) are not in Aeolian, so they are said
// to come from the modes by name (docs/architecture/theory-conventions.md, 3).
// The ♯11 of B♭(♯11) (E) is not looked at, so it keeps the modes of B♭; the
// bass of F/E♭, a minor seventh above the root, is (E♭ and A: Dorian only).
func TestBorrowedInCMajor(t *testing.T) {
	b := func(minor bool, modes ...material.Mode) want {
		return want{kind: reading.Borrowed, modes: modes, minor: minor}
	}
	check(t, major(cNat), []row{
		{"Fm (iv)", chord(fNat, min, ""), b(true, material.Phrygian, material.Aeolian, material.Locrian)},
		{"B♭ (♭VII)", chord(bFlat, maj, ""), b(true, material.Dorian, material.Mixolydian, material.Aeolian)},
		{"A♭ (♭VI)", chord(aFlat, maj, ""), b(true, material.Phrygian, material.Aeolian, material.Locrian)},
		{"E♭ (♭III)", chord(eFlat, maj, ""), b(true, material.Dorian, material.Phrygian, material.Aeolian)},
		{"Cm (i)", chord(cNat, min, ""), b(true, material.Dorian, material.Phrygian, material.Aeolian)},
		{"B♭maj7 (♭VIImaj7)", chord(bFlat, maj7, ""), b(false, material.Dorian, material.Mixolydian)},
		{"D♭ (♭II)", chord(dFlat, maj, ""), b(false, material.Phrygian, material.Locrian)},
		{"B♭(♯11)", chord(bFlat, "M3 P5 A11", ""), b(true, material.Dorian, material.Mixolydian, material.Aeolian)},
		{"F/E♭", chord(fNat, maj, "m7"), b(false, material.Dorian)},
	})
}

// Secondary dominants are named by shape and point to a degree; they are not
// borrowed (docs/architecture/theory-conventions.md, 2 and 3; Hutchinson 17.4,
// https://musictheory.pugetsound.edu/mt21c/AnalyzingSecondaryDominants.html).
// They carry no T, S or D (#102). C7 is V7/IV, not I7 (#101 on #83).
func TestSecondaryDominantsInCMajor(t *testing.T) {
	s := func(target int) want { return want{kind: reading.SecondaryDominant, target: target} }
	check(t, major(cNat), []row{
		{"D7 (V7/V)", chord(dNat, dom7, ""), s(5)},
		{"A7 (V7/ii)", chord(aNat, dom7, ""), s(2)},
		{"E (V/vi)", chord(eNat, maj, ""), s(6)},
		{"C7 (V7/IV)", chord(cNat, dom7, ""), s(4)},
		{"C/B♭ (V7/IV)", chord(cNat, maj, "m7"), s(4)},
	})
}

// Chords given no name (#101, D1; #89). B♯ major (♯VII) sounds as C, A♯ major
// (♯VI) as B♭ and F♭m (♭iv) as Em, but no mode on C spells A♯, B♯ or F♭
// (the owner's call on #102). E♭ in F♯ major and C𝄪
// major in C have no position: their roots are two half steps from the degree
// (E♯ and C), so they are read as facts only too (#102).
func TestFactsOnly(t *testing.T) {
	none := want{kind: reading.FactsOnly}
	check(t, major(cNat), []row{
		{"B♯ in C", chord(note(material.B, sharp), maj, ""), none},
		{"A♯ in C", chord(note(material.A, sharp), maj, ""), none},
		{"F♭m in C", chord(note(material.F, flat), min, ""), none},
		{"C𝄪 in C", chord(note(material.C, 2), maj, ""), none},
	})
	check(t, major(note(material.F, sharp)), []row{
		{"E♭ in F♯", chord(eFlat, maj, ""), none},
	})
}

// V7 in the twelve major keys is diatonic and dominant: the fifth degree of
// each major scale (Hutchinson 2.3,
// https://musictheory.pugetsound.edu/mt21c/MajorKeySignatures.html) with a
// major triad and a minor seventh.
func TestDominantSeventhInMajorKeys(t *testing.T) {
	tests := []struct {
		tonic, dominant material.NoteName
	}{
		{cNat, gNat},
		{gNat, dNat},
		{dNat, aNat},
		{aNat, eNat},
		{eNat, bNat},
		{bNat, note(material.F, sharp)},
		{note(material.F, sharp), note(material.C, sharp)},
		{dFlat, aFlat},
		{aFlat, eFlat},
		{eFlat, bFlat},
		{bFlat, fNat},
		{fNat, cNat},
	}
	for _, tt := range tests {
		check(t, major(tt.tonic), []row{
			{tt.tonic.String() + ": " + tt.dominant.String() + "7", chord(tt.dominant, dom7, ""),
				want{kind: reading.Diatonic, function: reading.Dominant}},
		})
	}
}

// The modes are compared by the interval from the tonic, so a mode that cannot
// be spelled on the tonic still holds a chord (decided in #102, not by the
// owner). On G𝄫, Dorian and Aeolian would need B triple flat, yet F𝄫 major
// (♭VII: F𝄫 A𝄫 C𝄫, a minor seventh, a major second and a perfect fourth above
// the tonic) is held by the same modes as B♭ in C.
func TestBorrowedOnUnspellableModes(t *testing.T) {
	check(t, major(note(material.G, -2)), []row{
		{"F𝄫 in G𝄫", chord(note(material.F, -2), maj, ""), want{
			kind: reading.Borrowed, minor: true,
			modes: []material.Mode{material.Dorian, material.Mixolydian, material.Aeolian},
		}},
	})
}

// A key other than Ionian is refused, as the position is (relation).
func TestReadChordOtherThanMajor(t *testing.T) {
	for m := material.Dorian; m <= material.Locrian; m++ {
		key, err := material.NewScale(cNat, m)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := reading.ReadChord(chord(gNat, dom7, ""), key); !errors.Is(err, relation.ErrNotMajor) {
			t.Errorf("ReadChord in C %s: err = %v; want ErrNotMajor", m, err)
		}
	}
}
