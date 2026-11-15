package notation_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
)

// ivs reads intervals such as "M3 P5 m7" for a test table.
func ivs(s string) []material.Interval {
	qualities := map[byte]material.Quality{
		'P': material.Perfect, 'M': material.Major, 'm': material.Minor, 'A': material.Augmented, 'd': material.Diminished,
	}
	var out []material.Interval
	for _, f := range strings.Fields(s) {
		number, err := strconv.Atoi(f[1:])
		if err != nil {
			panic(err)
		}
		i, err := material.NewInterval(number, qualities[f[0]])
		if err != nil {
			panic(err)
		}
		out = append(out, i)
	}
	return out
}

// chord builds a chord for a test table, like note. A bass interval of ""
// means no bass.
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

// The names are the display forms of docs/architecture/theory-conventions.md,
// 6 (the owner's decision B1 in Issue #78). The members of each kind follow the
// vocabulary of the material package (TestChordVocabulary and its sources).
func TestChordNames(t *testing.T) {
	c := note(material.C, 0)
	tests := []struct {
		name      string
		root      material.NoteName
		intervals string
		bass      string
	}{
		{"C", c, "M3 P5", ""},
		{"Cm", c, "m3 P5", ""},
		{"Cdim", c, "m3 d5", ""},
		{"Caug", c, "M3 A5", ""},
		{"C7", c, "M3 P5 m7", ""},
		{"Cmaj7", c, "M3 P5 M7", ""},
		{"Cm7", c, "m3 P5 m7", ""},
		{"Cm7(♭5)", c, "m3 d5 m7", ""},
		{"Cdim7", c, "m3 d5 d7", ""},
		{"Csus2", c, "M2 P5", ""},
		{"Csus4", c, "P4 P5", ""},
		{"C7sus4", c, "P4 P5 m7", ""},
		{"Cadd9", c, "M3 P5 M9", ""},
		{"C7(9)", c, "M3 P5 m7 M9", ""},
		{"Cmaj7(9)", c, "M3 P5 M7 M9", ""},
		{"Cm7(9)", c, "m3 P5 m7 M9", ""},
		{"C7(♭9)", c, "M3 P5 m7 m9", ""},
		{"C7(♯9)", c, "M3 P5 m7 A9", ""},
		{"C7(♯11)", c, "M3 P5 m7 A11", ""},
		{"C7(♭13)", c, "M3 P5 m7 m13", ""},
		{"C7(9,13)", c, "M3 P5 m7 M9 M13", ""},
		{"Cm7(9,11)", c, "m3 P5 m7 M9 P11", ""},
		{"F♯m7(♭5)", note(material.F, 1), "m3 d5 m7", ""},
		{"E♭maj7", note(material.E, -1), "M3 P5 M7", ""},
		{"B𝄫m", note(material.B, -2), "m3 P5", ""},
		{"C/E", c, "M3 P5", "M3"},
		{"C/D", c, "M3 P5", "M2"},
		{"G7/F", note(material.G, 0), "M3 P5 m7", "m7"},
		{"E♭/B♭", note(material.E, -1), "M3 P5", "P5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := chord(tt.root, tt.intervals, tt.bass)
			if got, err := notation.FormatChord(want); err != nil || got != tt.name {
				t.Errorf("FormatChord(%s %s / %s) = %q, %v; want %q", tt.root, tt.intervals, tt.bass, got, err, tt.name)
			}
			if got, err := notation.ParseChord(tt.name); err != nil || got != want {
				t.Errorf("ParseChord(%q) = %v %v, %v; want %v %v", tt.name, got.Root(), got.Intervals(), err, want.Root(), want.Intervals())
			}
		})
	}
}

// The aliases are the owner's decision A1 in Issue #78
// (docs/architecture/theory-conventions.md, 6).
func TestChordAliases(t *testing.T) {
	tests := []struct{ in, want string }{
		{"CM7", "Cmaj7"},
		{"CΔ7", "Cmaj7"},
		{"CM7(9)", "Cmaj7(9)"},
		{"Cmin", "Cm"},
		{"Cmin7", "Cm7"},
		{"Cmin7(9)", "Cm7(9)"},
		{"C°", "Cdim"},
		{"C°7", "Cdim7"},
		{"C+", "Caug"},
		{"Cm7♭5", "Cm7(♭5)"},
		{"Cm7b5", "Cm7(♭5)"},
		{"Cm7(b5)", "Cm7(♭5)"},
		{"Cø7", "Cm7(♭5)"},
		{"Cmin7♭5", "Cm7(♭5)"},
		{"C7(b9)", "C7(♭9)"},
		{"C7(#9)", "C7(♯9)"},
		{"C7(#11)", "C7(♯11)"},
		{"C7(9,b13)", "C7(9,♭13)"},
		{"F#m", "F♯m"},
		{"Bbmaj7", "B♭maj7"},
		{"Cxm", "C𝄪m"},
		{"Bbbm", "B𝄫m"},
		{"C/Bb", "C/B♭"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			c, err := notation.ParseChord(tt.in)
			if err != nil {
				t.Fatalf("ParseChord(%q): %v", tt.in, err)
			}
			if got, err := notation.FormatChord(c); err != nil || got != tt.want {
				t.Errorf("FormatChord(ParseChord(%q)) = %q, %v; want %q", tt.in, got, err, tt.want)
			}
		})
	}
}

func TestParseChordRejects(t *testing.T) {
	tests := []struct {
		in   string
		want error
		rest string // where reading stopped
	}{
		{"", notation.ErrNoteName, ""},
		{"cm", notation.ErrNoteName, "cm"},
		{"C-", notation.ErrChordKind, "-"},
		{"C-7", notation.ErrChordKind, "-7"},
		{"CM", notation.ErrChordKind, "M"},
		{"Cmaj", notation.ErrChordKind, "maj"},
		{"Csus", notation.ErrChordKind, "sus"},
		{"Cø", notation.ErrChordKind, "ø"},
		{"C+7", notation.ErrChordKind, "+7"},
		{"C6", notation.ErrChordKind, "6"},
		// 9 chords are written C7(9) (decision B1); C9 is not an alias.
		{"C9", notation.ErrChordKind, "9"},
		{"Cmaj9", notation.ErrChordKind, "maj9"},
		{"Cm9", notation.ErrChordKind, "m9"},
		{"C###", notation.ErrAccidental, "###"},
		{"C♯♯♯m", notation.ErrAccidental, "♯♯♯m"},
		{"C7(9", notation.ErrUnclosedParen, "(9"},
		{"Cm7(♭5", notation.ErrUnclosedParen, "(♭5"},
		{"C(9)", notation.ErrNoTensions, "(9)"},
		{"Csus4(9)", notation.ErrNoTensions, "(9)"},
		{"Cm7(♭5)(9)", notation.ErrNoTensions, "(9)"},
		{"C7()", notation.ErrTension, ")"},
		{"C7(9,)", notation.ErrTension, ")"},
		{"C7(♭11)", notation.ErrTension, "♭11)"},
		{"C7(♯13)", notation.ErrTension, "♯13)"},
		{"C7(♭♭9)", notation.ErrTension, "♭♭9)"},
		{"Cm7(♭5,9)", notation.ErrTension, "♭5,9)"},
		{"C7(9, 13)", notation.ErrTension, " 13)"},
		{"C7(13,9)", notation.ErrTensionOrder, "9)"},
		{"C7(9,♭9)", notation.ErrTensionOrder, "♭9)"},
		{"C7(9)x", notation.ErrTrailing, "x"},
		{"C/", notation.ErrNoteName, ""},
		{"C/e", notation.ErrNoteName, "e"},
		{"C/E7", notation.ErrTrailing, "7"},
		{"C/E/G", notation.ErrTrailing, "/G"},
		// No interval from C♯ up to E𝄫 (a third of one semitone).
		{"C♯/E𝄫", material.ErrInvalidInterval, "E𝄫"},
		// The third of B𝄪 would be beyond a double sharp.
		{"B𝄪", material.ErrAlterOutOfRange, "B𝄪"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := notation.ParseChord(tt.in)
			var pe *notation.ParseError
			if !errors.As(err, &pe) || !errors.Is(err, tt.want) || pe.Rest != tt.rest || pe.Input != tt.in {
				t.Errorf("ParseChord(%q) = %v, %v; want %v at %q", tt.in, got.Intervals(), err, tt.want, tt.rest)
			}
		})
	}
}

func TestFormatChordNoName(t *testing.T) {
	c := note(material.C, 0)
	tests := []struct {
		name      string
		intervals string
	}{
		{"the root alone", ""},
		{"a fifth alone (C5)", "P5"},
		{"a sixth (C6)", "M3 P5 M6"},
		{"a third without a fifth", "M3 m7"},
		{"a minor triad with a major seventh", "m3 P5 M7"},
		{"a minor ninth without a seventh", "M3 P5 m9"},
		{"a ninth on a minor triad without a seventh", "m3 P5 M9"},
		{"a tension on 7sus4", "P4 P5 m7 M9"},
		{"a tension on m7(♭5)", "m3 d5 m7 M9"},
		{"a diminished eleventh", "M3 P5 m7 d11"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := notation.FormatChord(chord(c, tt.intervals, "")); !errors.Is(err, notation.ErrNoName) {
				t.Errorf("FormatChord(C %s) = %q, %v; want error %v", tt.intervals, got, err, notation.ErrNoName)
			}
		})
	}
}

// The rule and the examples are docs/architecture/theory-conventions.md, 6.
func TestFormatChordForDisplay(t *testing.T) {
	tests := []struct {
		root            material.NoteName
		intervals, bass string
		spelled, shown  string
	}{
		{note(material.B, -2), "M3 P5", "", "B𝄫", "A"},
		{note(material.C, -1), "m3 P5 m7", "", "C♭m7", "Bm7"},
		{note(material.F, -1), "M3 P5 M7", "", "F♭maj7", "Emaj7"},
		{note(material.F, 2), "M3 P5 m7 M9", "", "F𝄪7(9)", "G7(9)"},
		{note(material.C, 0), "M3 P5", "d4", "C/F♭", "C/E"},
		{note(material.E, 1), "M3 P5", "", "E♯", "E♯"},
		// The root and the bass are respelled each on its own.
		{note(material.B, -2), "M3 P5", "M3", "B𝄫/D♭", "A/D♭"},
	}
	for _, tt := range tests {
		t.Run(tt.spelled, func(t *testing.T) {
			c := chord(tt.root, tt.intervals, tt.bass)
			if got, err := notation.FormatChord(c); err != nil || got != tt.spelled {
				t.Errorf("FormatChord = %q, %v; want %q", got, err, tt.spelled)
			}
			got, err := notation.FormatChordForDisplay(c)
			if err != nil || got != tt.shown {
				t.Fatalf("FormatChordForDisplay = %q, %v; want %q", got, err, tt.shown)
			}
			if back, err := notation.ParseChord(got); err != nil || back.SoundingPitchClasses() != c.SoundingPitchClasses() {
				t.Errorf("ParseChord(%q) sounds %s, %v; want %s", got, back.SoundingPitchClasses(), err, c.SoundingPitchClasses())
			}
		})
	}
}

// displayNames returns every chord name in the display form for the root and
// bass spellings given: each kind, and 7, maj7 and m7 with each set of
// tensions, with and without a bass.
func displayNames(root string, basses []string) []string {
	kinds := []string{"", "m", "dim", "aug", "7", "maj7", "m7", "m7(♭5)", "dim7", "sus2", "sus4", "7sus4", "add9"}
	for _, k := range []string{"7", "maj7", "m7"} {
		for _, n9 := range []string{"", "9", "♭9", "♯9"} {
			for _, n11 := range []string{"", "11", "♯11"} {
				for _, n13 := range []string{"", "13", "♭13"} {
					var ts []string
					for _, t := range []string{n9, n11, n13} {
						if t != "" {
							ts = append(ts, t)
						}
					}
					if len(ts) > 0 {
						kinds = append(kinds, k+"("+strings.Join(ts, ",")+")")
					}
				}
			}
		}
	}
	var names []string
	for _, k := range kinds {
		names = append(names, root+k)
		for _, b := range basses {
			names = append(names, root+k+"/"+b)
		}
	}
	return names
}

// Reading a name in the display form and writing it back gives the same text,
// and reading that gives the same chord. Writing for display keeps the
// sounding pitch classes. The only failures allowed are a member or bass beyond
// a double sharp or flat, and a bass with no interval to the root.
func TestChordRoundTrip(t *testing.T) {
	var spellings []string
	for _, n := range allNotes() {
		spellings = append(spellings, notation.FormatNoteName(n))
	}
	read := 0
	for _, root := range spellings {
		for _, name := range displayNames(root, spellings) {
			c, err := notation.ParseChord(name)
			if errors.Is(err, material.ErrAlterOutOfRange) || errors.Is(err, material.ErrInvalidInterval) {
				continue
			}
			if err != nil {
				t.Fatalf("ParseChord(%q): %v", name, err)
			}
			read++
			written, err := notation.FormatChord(c)
			if err != nil || written != name {
				t.Fatalf("FormatChord(ParseChord(%q)) = %q, %v", name, written, err)
			}
			if back, err := notation.ParseChord(written); err != nil || back != c {
				t.Fatalf("ParseChord(%q) = %v, %v; want the chord read from %q", written, back.Intervals(), err, name)
			}
			shown, err := notation.FormatChordForDisplay(c)
			if err != nil {
				t.Fatalf("FormatChordForDisplay(ParseChord(%q)): %v", name, err)
			}
			if back, err := notation.ParseChord(shown); err != nil || back.SoundingPitchClasses() != c.SoundingPitchClasses() {
				t.Fatalf("ParseChord(%q) = %s, %v; want the sounding pitch classes %s of %q", shown, back.SoundingPitchClasses(), err, c.SoundingPitchClasses(), name)
			}
		}
	}
	t.Logf("%d names read", read)
	if read == 0 {
		t.Fatal("no name was read")
	}
}
