package notation_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
)

// position builds a position for a test table. A target of 0 means none.
func position(degree int, alter material.Alteration, intervals string, target int) material.Position {
	p, err := material.NewPosition(degree, alter, material.ShapeOf(chord(material.NoteName{}, intervals, "")))
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
func describePosition(p material.Position) string {
	target, _ := p.Target()
	return fmt.Sprintf("degree %d alter %d %v target %d", p.Degree(), p.Alteration(), p.Shape().Intervals(), target)
}

// The forms are from docs/architecture/theory-conventions.md, 6. The owner's
// decision C1 in Issue #78: uppercase for major and augmented, lowercase for
// minor and diminished, °, ø7 and +, an accidental before an altered degree,
// the kinds and tensions of chord names, and "/" with the target. A convention
// of our own: a chord without a third takes its suffix on the degree, in the
// case of the triad of its degree in the major key (I ii iii IV V vi vii°),
// and in uppercase on an altered degree (the owner's decision in #82). Decided in the implementation (#80): the target is one of
// ii, iii, IV, V and vi, and the degree before "/" is counted from the root of
// the target. The chords in C major are the major scale C D E F G A B with the
// members of each kind.
func TestRomanNumerals(t *testing.T) {
	n, flat, sharp := material.Natural, material.Flat, material.Sharp
	tests := []struct {
		roman     string
		degree    int
		alter     material.Alteration
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
		{"♭VIIsus4", 7, flat, "P4 P5", 0, "B♭sus4"},
		{"♭IIIsus2", 3, flat, "M2 P5", 0, "E♭sus2"},
		{"♭VIsus4", 6, flat, "P4 P5", 0, "A♭sus4"},
		{"♭VII7sus4", 7, flat, "P4 P5 m7", 0, "B♭7sus4"},
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

// The chords of I, IV, V7, vi, ♭VII and ♭VI in the twelve major keys of the
// app (docs/architecture/engine.md) and G♭. The notes are those of each major
// scale; ♭VII and ♭VI lower the degree a half step with the same letter. In a
// key the chord names are not respelled: FormatChord keeps the letter of the
// degree (on D♭, ♭VII is C♭ and ♭VI is B𝄫; the owner's decision in #82,
// docs/architecture/theory-conventions.md, 6).
func TestRomanInMajorKeys(t *testing.T) {
	romans := [6]string{"I", "IV", "V7", "vi", "♭VII", "♭VI"}
	tests := []struct {
		tonic   string
		spelled [6]string
	}{
		{"C", [6]string{"C", "F", "G7", "Am", "B♭", "A♭"}},
		{"G", [6]string{"G", "C", "D7", "Em", "F", "E♭"}},
		{"D", [6]string{"D", "G", "A7", "Bm", "C", "B♭"}},
		{"A", [6]string{"A", "D", "E7", "F♯m", "G", "F"}},
		{"E", [6]string{"E", "A", "B7", "C♯m", "D", "C"}},
		{"B", [6]string{"B", "E", "F♯7", "G♯m", "A", "G"}},
		{"F♯", [6]string{"F♯", "B", "C♯7", "D♯m", "E", "D"}},
		{"G♭", [6]string{"G♭", "C♭", "D♭7", "E♭m", "F♭", "E𝄫"}},
		{"D♭", [6]string{"D♭", "G♭", "A♭7", "B♭m", "C♭", "B𝄫"}},
		{"A♭", [6]string{"A♭", "D♭", "E♭7", "Fm", "G♭", "F♭"}},
		{"E♭", [6]string{"E♭", "A♭", "B♭7", "Cm", "D♭", "C♭"}},
		{"B♭", [6]string{"B♭", "E♭", "F7", "Gm", "A♭", "G♭"}},
		{"F", [6]string{"F", "B♭", "C7", "Dm", "E♭", "D♭"}},
	}
	for _, tt := range tests {
		tonic, err := notation.ParseNoteName(tt.tonic)
		if err != nil {
			t.Fatalf("ParseNoteName(%q): %v", tt.tonic, err)
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
			})
		}
	}
}

// Every position on the shapes of the vocabulary (each kind, and 7, maj7 and
// m7 with each set of tensions), on every degree and alteration, with and
// without each target, is written and read back to the same position, and the
// text read back is written the same.
func TestRomanRoundTrip(t *testing.T) {
	var shapes []material.Shape
	for _, name := range displayNames("C", nil) {
		c, err := notation.ParseChord(name)
		if err != nil {
			t.Fatalf("ParseChord(%q): %v", name, err)
		}
		shapes = append(shapes, material.ShapeOf(c))
	}
	for _, shape := range shapes {
		for degree := 1; degree <= 7; degree++ {
			for _, alter := range []material.Alteration{material.Flat, material.Natural, material.Sharp} {
				for _, target := range []int{0, 2, 3, 4, 5, 6} {
					p, err := material.NewPosition(degree, alter, shape)
					if err == nil && target != 0 {
						p, err = p.WithTarget(target)
					}
					if err != nil {
						t.Fatalf("degree %d alter %d target %d: %v", degree, alter, target, err)
					}
					s, err := notation.FormatRoman(p)
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
		// No third: the case of the triad of the degree, and uppercase on an
		// altered degree.
		{"vsus4", notation.ErrRomanCase, "vsus4"},
		{"IIsus2", notation.ErrRomanCase, "IIsus2"},
		{"VIIsus4", notation.ErrRomanCase, "VIIsus4"},
		{"♭viisus4", notation.ErrRomanCase, "viisus4"},
		{"♯ivsus4", notation.ErrRomanCase, "ivsus4"},
		// One target: ii, iii, IV, V or vi.
		{"V7/", material.ErrTarget, ""},
		{"V/I", material.ErrTarget, "I"},
		{"V/vii", material.ErrTarget, "vii"},
		{"V/II", material.ErrTarget, "II"},
		{"V/iv", material.ErrTarget, "iv"},
		{"V/♭VII", material.ErrTarget, "♭VII"},
		{"V/B", material.ErrTarget, "B"},
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
	withBass, err := material.NewPosition(5, material.Natural, material.ShapeOf(chord(material.NoteName{}, "M3 P5", "M3")))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		p    material.Position
		want error
	}{
		{"a bass (V with the third in the bass)", withBass, notation.ErrRomanBass},
		{"the zero value (the root alone)", material.Position{}, notation.ErrNoName},
		{"a sixth (C6 on I)", position(1, material.Natural, "M3 P5 M6", 0), notation.ErrNoName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := notation.FormatRoman(tt.p); !errors.Is(err, tt.want) {
				t.Errorf("FormatRoman(%s) = %q, %v; want error %v", describePosition(tt.p), got, err, tt.want)
			}
		})
	}
}

// The bass of the shape comes with the chord: V with the third in the bass is
// G/B in C. The chord is checked in material; here the name is written.
func TestFormatChordOfPositionWithBass(t *testing.T) {
	withBass, err := material.NewPosition(5, material.Natural, material.ShapeOf(chord(material.NoteName{}, "M3 P5", "M3")))
	if err != nil {
		t.Fatal(err)
	}
	c, err := withBass.Chord(note(material.C, 0))
	if got, fmtErr := notation.FormatChord(c); err != nil || fmtErr != nil || got != "G/B" {
		t.Errorf("V with the third in the bass in C = %q, %v, %v; want G/B", got, err, fmtErr)
	}
}
