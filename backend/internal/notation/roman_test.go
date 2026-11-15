package notation_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
)

// position builds a position for a test table. A target of 0 means none.
func position(degree int, alter notation.Alteration, intervals string, target int) notation.Position {
	p, err := notation.NewPosition(degree, alter, notation.ShapeOf(chord(material.NoteName{}, intervals, "")))
	if err != nil {
		panic(err)
	}
	if target != 0 {
		if p, err = p.WithTarget(target); err != nil {
			panic(err)
		}
	}
	return p
}

// describePosition shows a position in a failure message.
func describePosition(p notation.Position) string {
	target, _ := p.Target()
	return fmt.Sprintf("degree %d alter %d %v target %d", p.Degree(), p.Alteration(), p.Shape().Intervals(), target)
}

// The forms are the owner's decision C1 in Issue #78
// (docs/architecture/theory-conventions.md, 6): uppercase for major and
// augmented, lowercase for minor and diminished, °, ø7 and +, an accidental
// before an altered degree, the kinds and tensions of chord names, and "/"
// with the target. A chord without a third takes the case of the triad of its
// degree in the major key (I ii iii IV V vi vii°). The chords in C major are
// the major scale C D E F G A B with the members of each kind.
func TestRomanNumerals(t *testing.T) {
	n, flat, sharp := notation.Natural, notation.Flat, notation.Sharp
	tests := []struct {
		roman     string
		degree    int
		alter     notation.Alteration
		intervals string
		target    int
		inC       string
	}{
		{"I", 1, n, "M3 P5", 0, "C"},
		{"ii", 2, n, "m3 P5", 0, "Dm"},
		{"iii", 3, n, "m3 P5", 0, "Em"},
		{"IV", 4, n, "M3 P5", 0, "F"},
		{"V", 5, n, "M3 P5", 0, "G"},
		{"vi", 6, n, "m3 P5", 0, "Am"},
		{"vii°", 7, n, "m3 d5", 0, "Bdim"},
		{"Imaj7", 1, n, "M3 P5 M7", 0, "Cmaj7"},
		{"ii7", 2, n, "m3 P5 m7", 0, "Dm7"},
		{"iii7", 3, n, "m3 P5 m7", 0, "Em7"},
		{"IVmaj7", 4, n, "M3 P5 M7", 0, "Fmaj7"},
		{"V7", 5, n, "M3 P5 m7", 0, "G7"},
		{"vi7", 6, n, "m3 P5 m7", 0, "Am7"},
		{"viiø7", 7, n, "m3 d5 m7", 0, "Bm7(♭5)"},
		{"vii°7", 7, n, "m3 d5 d7", 0, "Bdim7"},
		{"I7", 1, n, "M3 P5 m7", 0, "C7"},
		// Chords from outside the key: an accidental before the degree.
		{"♭VII", 7, flat, "M3 P5", 0, "B♭"},
		{"♭VI", 6, flat, "M3 P5", 0, "A♭"},
		{"♭III", 3, flat, "M3 P5", 0, "E♭"},
		{"♭II", 2, flat, "M3 P5", 0, "D♭"},
		{"♭III+", 3, flat, "M3 A5", 0, "E♭aug"},
		{"♯iv°", 4, sharp, "m3 d5", 0, "F♯dim"},
		{"iv", 4, n, "m3 P5", 0, "Fm"},
		// Secondary chords: the degree is counted from the root of the target.
		{"V7/V", 5, n, "M3 P5 m7", 5, "D7"},
		{"V7/vi", 5, n, "M3 P5 m7", 6, "E7"},
		{"V/ii", 5, n, "M3 P5", 2, "A"},
		{"V7/iii", 5, n, "M3 P5 m7", 3, "B7"},
		{"V7/IV", 5, n, "M3 P5 m7", 4, "C7"},
		{"vii°7/V", 7, n, "m3 d5 d7", 5, "F♯dim7"},
		// No third: the case of the triad of the degree.
		{"Vsus4", 5, n, "P4 P5", 0, "Gsus4"},
		{"V7sus4", 5, n, "P4 P5 m7", 0, "G7sus4"},
		{"iisus2", 2, n, "M2 P5", 0, "Dsus2"},
		{"iiisus4", 3, n, "P4 P5", 0, "Esus4"},
		{"Isus2", 1, n, "M2 P5", 0, "Csus2"},
		// Tensions as in chord names.
		{"V7(9)", 5, n, "M3 P5 m7 M9", 0, "G7(9)"},
		{"V7(♭9,♭13)", 5, n, "M3 P5 m7 m9 m13", 0, "G7(♭9,♭13)"},
		{"Imaj7(9)", 1, n, "M3 P5 M7 M9", 0, "Cmaj7(9)"},
		{"ii7(9,11)", 2, n, "m3 P5 m7 M9 P11", 0, "Dm7(9,11)"},
		{"Iadd9", 1, n, "M3 P5 M9", 0, "Cadd9"},
	}
	for _, tt := range tests {
		t.Run(tt.roman, func(t *testing.T) {
			want := position(tt.degree, tt.alter, tt.intervals, tt.target)
			if got, err := notation.FormatRoman(want); err != nil || got != tt.roman {
				t.Errorf("FormatRoman(%s) = %q, %v; want %q", describePosition(want), got, err, tt.roman)
			}
			got, err := notation.ParseRoman(tt.roman)
			if err != nil || got != want {
				t.Fatalf("ParseRoman(%q) = %s, %v; want %s", tt.roman, describePosition(got), err, describePosition(want))
			}
			c, err := got.Chord(note(material.C, 0))
			if err != nil {
				t.Fatalf("%q in C: %v", tt.roman, err)
			}
			if name, err := notation.FormatChord(c); err != nil || name != tt.inC {
				t.Errorf("%q in C = %q, %v; want %q", tt.roman, name, err, tt.inC)
			}
		})
	}
}

// The chords of I, IV, V7, vi and ♭VII in the twelve major keys (F♯ and G♭
// both). The notes are those of each major scale; ♭VII lowers the seventh
// degree a half step with the same letter. The chord names keep the spelling
// (FormatChord); the shown names respell C♭ and F♭
// (docs/architecture/theory-conventions.md, 6).
func TestRomanInMajorKeys(t *testing.T) {
	romans := [5]string{"I", "IV", "V7", "vi", "♭VII"}
	tests := []struct {
		tonic   string
		spelled [5]string
		shown   [5]string // when it differs from spelled
	}{
		{"C", [5]string{"C", "F", "G7", "Am", "B♭"}, [5]string{}},
		{"G", [5]string{"G", "C", "D7", "Em", "F"}, [5]string{}},
		{"D", [5]string{"D", "G", "A7", "Bm", "C"}, [5]string{}},
		{"A", [5]string{"A", "D", "E7", "F♯m", "G"}, [5]string{}},
		{"E", [5]string{"E", "A", "B7", "C♯m", "D"}, [5]string{}},
		{"B", [5]string{"B", "E", "F♯7", "G♯m", "A"}, [5]string{}},
		{"F♯", [5]string{"F♯", "B", "C♯7", "D♯m", "E"}, [5]string{}},
		{"G♭", [5]string{"G♭", "C♭", "D♭7", "E♭m", "F♭"}, [5]string{"G♭", "B", "D♭7", "E♭m", "E"}},
		{"D♭", [5]string{"D♭", "G♭", "A♭7", "B♭m", "C♭"}, [5]string{"D♭", "G♭", "A♭7", "B♭m", "B"}},
		{"A♭", [5]string{"A♭", "D♭", "E♭7", "Fm", "G♭"}, [5]string{}},
		{"E♭", [5]string{"E♭", "A♭", "B♭7", "Cm", "D♭"}, [5]string{}},
		{"B♭", [5]string{"B♭", "E♭", "F7", "Gm", "A♭"}, [5]string{}},
		{"F", [5]string{"F", "B♭", "C7", "Dm", "E♭"}, [5]string{}},
	}
	for _, tt := range tests {
		tonic, err := notation.ParseNoteName(tt.tonic)
		if err != nil {
			t.Fatalf("ParseNoteName(%q): %v", tt.tonic, err)
		}
		shown := tt.shown
		if shown == [5]string{} {
			shown = tt.spelled
		}
		for i, roman := range romans {
			t.Run(tt.tonic+" "+roman, func(t *testing.T) {
				p, err := notation.ParseRoman(roman)
				if err != nil {
					t.Fatalf("ParseRoman(%q): %v", roman, err)
				}
				c, err := p.Chord(tonic)
				if err != nil {
					t.Fatalf("%s in %s: %v", roman, tt.tonic, err)
				}
				if got, err := notation.FormatChord(c); err != nil || got != tt.spelled[i] {
					t.Errorf("FormatChord(%s in %s) = %q, %v; want %q", roman, tt.tonic, got, err, tt.spelled[i])
				}
				if got, err := notation.FormatChordForDisplay(c); err != nil || got != shown[i] {
					t.Errorf("FormatChordForDisplay(%s in %s) = %q, %v; want %q", roman, tt.tonic, got, err, shown[i])
				}
			})
		}
	}
}

// Every position on the shapes of the vocabulary (each kind, and 7, maj7 and
// m7 with each set of tensions), on every degree and alteration, with and
// without each target, is written and read back to the same position, and the
// text read back is written the same. The one failure allowed is a chord
// without a third on an altered degree (ErrAlteredSus), and it must fail.
func TestRomanRoundTrip(t *testing.T) {
	var shapes []notation.Shape
	for _, name := range displayNames("C", nil) {
		c, err := notation.ParseChord(name)
		if err != nil {
			t.Fatalf("ParseChord(%q): %v", name, err)
		}
		shapes = append(shapes, notation.ShapeOf(c))
	}
	for _, shape := range shapes {
		onC, err := shape.On(material.NoteName{})
		if err != nil {
			t.Fatalf("shape %v on C: %v", shape.Intervals(), err)
		}
		noThird := onC.Triad() == material.TriadSus4 || onC.Triad() == material.TriadSus2
		for degree := 1; degree <= 7; degree++ {
			for _, alter := range []notation.Alteration{notation.Flat, notation.Natural, notation.Sharp} {
				for _, target := range []int{0, 2, 3, 4, 5, 6} {
					p, err := notation.NewPosition(degree, alter, shape)
					if err == nil && target != 0 {
						p, err = p.WithTarget(target)
					}
					if err != nil {
						t.Fatalf("degree %d alter %d target %d: %v", degree, alter, target, err)
					}
					s, err := notation.FormatRoman(p)
					if noThird && alter != notation.Natural {
						if !errors.Is(err, notation.ErrAlteredSus) {
							t.Fatalf("FormatRoman(%s) = %q, %v; want error %v", describePosition(p), s, err, notation.ErrAlteredSus)
						}
						continue
					}
					if err != nil {
						t.Fatalf("FormatRoman(%s): %v", describePosition(p), err)
					}
					back, err := notation.ParseRoman(s)
					if err != nil || back != p {
						t.Fatalf("ParseRoman(%q) = %s, %v; want %s", s, describePosition(back), err, describePosition(p))
					}
					if again, err := notation.FormatRoman(back); err != nil || again != s {
						t.Fatalf("FormatRoman(ParseRoman(%q)) = %q, %v", s, again, err)
					}
				}
			}
		}
	}
}

func TestParseRomanRejects(t *testing.T) {
	tests := []struct {
		in   string
		want error
		rest string // where reading stopped
	}{
		{"", notation.ErrNumeral, ""},
		{" I", notation.ErrNumeral, " I"},
		{"C", notation.ErrNumeral, "C"},
		// Only the display accidentals ♭ and ♯, and only one.
		{"bVII", notation.ErrNumeral, "bVII"},
		{"#iv°", notation.ErrNumeral, "#iv°"},
		{"♭♭VII", notation.ErrNumeral, "♭VII"},
		{"𝄫VII", notation.ErrNumeral, "𝄫VII"},
		// The case must be one for the whole numeral, and match the kind.
		{"Iv", notation.ErrRomanKind, "v"},
		{"VIII", notation.ErrRomanKind, "I"},
		{"iimaj7", notation.ErrRomanKind, "maj7"},
		{"I°", notation.ErrRomanKind, "°"},
		{"i+", notation.ErrRomanKind, "+"},
		{"Iø7", notation.ErrRomanKind, "ø7"},
		{"viiø", notation.ErrRomanKind, "ø"},
		// The display forms only: not the aliases of chord names or RomanText.
		{"Vm", notation.ErrRomanKind, "m"},
		{"V-", notation.ErrRomanKind, "-"},
		{"viio", notation.ErrRomanKind, "o"},
		{"IM7", notation.ErrRomanKind, "M7"},
		{"IΔ7", notation.ErrRomanKind, "Δ7"},
		{"V9", notation.ErrRomanKind, "9"},
		{"V7x", notation.ErrRomanKind, "7x"},
		{"V7(b9)", notation.ErrTension, "b9)"},
		{"V7(9,#11)", notation.ErrTension, "#11)"},
		// Inversions are not written.
		{"I6", notation.ErrRomanKind, "6"},
		{"V65", notation.ErrRomanKind, "65"},
		// Tensions as in chord names.
		{"V7(9", notation.ErrUnclosedParen, "(9"},
		{"V(9)", notation.ErrNoTensions, "(9)"},
		{"V7sus4(9)", notation.ErrNoTensions, "(9)"},
		{"viiø7(9)", notation.ErrNoTensions, "(9)"},
		{"V7(13,9)", notation.ErrTensionOrder, "9)"},
		// No third: the case of the triad of the degree, and no alteration.
		{"vsus4", notation.ErrRomanCase, "vsus4"},
		{"IIsus2", notation.ErrRomanCase, "IIsus2"},
		{"VIIsus4", notation.ErrRomanCase, "VIIsus4"},
		{"♭VIIsus4", notation.ErrAlteredSus, "♭VIIsus4"},
		// One target: ii, iii, IV, V or vi.
		{"V7/", notation.ErrTarget, ""},
		{"V/I", notation.ErrTarget, "I"},
		{"V/vii", notation.ErrTarget, "vii"},
		{"V/II", notation.ErrTarget, "II"},
		{"V/iv", notation.ErrTarget, "iv"},
		{"V/♭VII", notation.ErrTarget, "♭VII"},
		{"V/B", notation.ErrTarget, "B"},
		{"V7/V/V", notation.ErrTrailing, "/V"},
		{"V7/Vx", notation.ErrTrailing, "x"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := notation.ParseRoman(tt.in)
			var pe *notation.ParseError
			if !errors.As(err, &pe) || !errors.Is(err, tt.want) || pe.Rest != tt.rest || pe.Input != tt.in {
				t.Errorf("ParseRoman(%q) = %s, %v; want %v at %q", tt.in, describePosition(got), err, tt.want, tt.rest)
			}
		})
	}
}

func TestFormatRomanRejects(t *testing.T) {
	withBass, err := notation.NewPosition(5, notation.Natural, notation.ShapeOf(chord(material.NoteName{}, "M3 P5", "M3")))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		p    notation.Position
		want error
	}{
		{"a bass (V with the third in the bass)", withBass, notation.ErrRomanBass},
		{"the zero value (the root alone)", notation.Position{}, notation.ErrNoName},
		{"a sixth (C6 on I)", position(1, notation.Natural, "M3 P5 M6", 0), notation.ErrNoName},
		{"a chord without a third on ♭VII", position(7, notation.Flat, "P4 P5", 0), notation.ErrAlteredSus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := notation.FormatRoman(tt.p); !errors.Is(err, tt.want) {
				t.Errorf("FormatRoman(%s) = %q, %v; want error %v", describePosition(tt.p), got, err, tt.want)
			}
		})
	}
}

func TestNewPositionRejects(t *testing.T) {
	var shape notation.Shape
	tests := []struct {
		degree int
		alter  notation.Alteration
		want   error
	}{
		{0, notation.Natural, notation.ErrDegree},
		{8, notation.Natural, notation.ErrDegree},
		{1, notation.Alteration(2), notation.ErrAlteration},
		{1, notation.Alteration(-2), notation.ErrAlteration},
	}
	for _, tt := range tests {
		if _, err := notation.NewPosition(tt.degree, tt.alter, shape); !errors.Is(err, tt.want) {
			t.Errorf("NewPosition(%d, %d) = %v; want %v", tt.degree, tt.alter, err, tt.want)
		}
	}
	p, err := notation.NewPosition(5, notation.Natural, shape)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []int{0, 1, 7, 8} {
		if _, err := p.WithTarget(target); !errors.Is(err, notation.ErrTarget) {
			t.Errorf("WithTarget(%d) = %v; want %v", target, err, notation.ErrTarget)
		}
	}
}

// The bass of the shape comes with the chord: V with the third in the bass is
// G/B in C. A root, a member or the root of the target beyond a double sharp
// fails. In G♯ major the seventh degree is F𝄪, so ♯VII is beyond it. The
// augmented fifth of C𝄪 would be G with three sharps. The fifth degree of B𝄪
// major would be F with three sharps.
func TestPositionChord(t *testing.T) {
	withBass, err := notation.NewPosition(5, notation.Natural, notation.ShapeOf(chord(material.NoteName{}, "M3 P5", "M3")))
	if err != nil {
		t.Fatal(err)
	}
	c, err := withBass.Chord(note(material.C, 0))
	if got, fmtErr := notation.FormatChord(c); err != nil || fmtErr != nil || got != "G/B" {
		t.Errorf("V with the third in the bass in C = %q, %v, %v; want G/B", got, err, fmtErr)
	}
	tests := []struct {
		name  string
		tonic material.NoteName
		p     notation.Position
	}{
		{"♯VII in G♯ (beyond F𝄪)", note(material.G, 1), position(7, notation.Sharp, "M3 P5", 0)},
		{"I+ in C𝄪 (the fifth)", note(material.C, 2), position(1, notation.Natural, "M3 A5", 0)},
		{"V/V in B𝄪 (the target)", note(material.B, 2), position(5, notation.Natural, "M3 P5", 5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.p.Chord(tt.tonic); !errors.Is(err, material.ErrAlterOutOfRange) {
				t.Errorf("Chord(%s) = %s, %v; want error %v", tt.tonic, describe(got), err, material.ErrAlterOutOfRange)
			}
		})
	}
}
