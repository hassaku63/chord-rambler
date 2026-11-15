package notation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
)

// note builds a note name for a test table. The tables hold only valid
// values, so a failure here is a mistake in the table.
func note(l material.Letter, alter int) material.NoteName {
	n, err := material.NewNoteName(l, alter)
	if err != nil {
		panic(err)
	}
	return n
}

// allNotes returns the 35 note names: seven letters, double flat to double
// sharp.
func allNotes() []material.NoteName {
	var notes []material.NoteName
	for l := material.C; l <= material.B; l++ {
		for a := -2; a <= 2; a++ {
			notes = append(notes, note(l, a))
		}
	}
	return notes
}

// The spellings read are those of Issue #79 (docs/architecture/theory-conventions.md, 6).
func TestParseNoteName(t *testing.T) {
	tests := []struct {
		in    string
		want  material.NoteName
		write string
	}{
		{"C", note(material.C, 0), "C"},
		{"A", note(material.A, 0), "A"},
		{"F♯", note(material.F, 1), "F♯"},
		{"F#", note(material.F, 1), "F♯"},
		{"B♭", note(material.B, -1), "B♭"},
		{"Bb", note(material.B, -1), "B♭"},
		{"C𝄪", note(material.C, 2), "C𝄪"},
		{"Cx", note(material.C, 2), "C𝄪"},
		{"C##", note(material.C, 2), "C𝄪"},
		{"B𝄫", note(material.B, -2), "B𝄫"},
		{"Bbb", note(material.B, -2), "B𝄫"},
		{"B♭♭", note(material.B, -2), "B𝄫"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := notation.ParseNoteName(tt.in)
			if err != nil || got != tt.want {
				t.Fatalf("ParseNoteName(%q) = %s, %v; want %s", tt.in, got, err, tt.want)
			}
			if w := notation.FormatNoteName(got); w != tt.write {
				t.Errorf("FormatNoteName(%s) = %q, want %q", got, w, tt.write)
			}
		})
	}
}

func TestNoteNameRoundTrip(t *testing.T) {
	for _, n := range allNotes() {
		s := notation.FormatNoteName(n)
		if got, err := notation.ParseNoteName(s); err != nil || got != n {
			t.Errorf("ParseNoteName(FormatNoteName(%s) = %q) = %s, %v; want %s", n, s, got, err, n)
		}
	}
}

func TestParseNoteNameRejects(t *testing.T) {
	tests := []struct {
		in   string
		want error
		rest string // where reading stopped
	}{
		{"", notation.ErrNoteName, ""},
		{"c", notation.ErrNoteName, "c"},
		{"H", notation.ErrNoteName, "H"},
		{"C###", notation.ErrAccidental, "###"},
		{"C𝄪♯", notation.ErrAccidental, "𝄪♯"},
		{"Bbbb", notation.ErrAccidental, "bbb"},
		{"C#b", notation.ErrAccidental, "#b"},
		{"C7", notation.ErrTrailing, "7"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := notation.ParseNoteName(tt.in)
			var pe *notation.ParseError
			if !errors.As(err, &pe) || !errors.Is(err, tt.want) || pe.Rest != tt.rest || pe.Input != tt.in {
				t.Errorf("ParseNoteName(%q) = %s, %v; want %v at %q", tt.in, got, err, tt.want, tt.rest)
			}
		})
	}
}

// The rule and its three examples (B♭♭ is A, C♭ is B, F♭ is E) are
// docs/architecture/theory-conventions.md, 6. The other rows move to the
// neighbouring letter in the same way.
func TestDisplaySpelling(t *testing.T) {
	tests := []struct {
		in, want material.NoteName
	}{
		{note(material.B, -2), note(material.A, 0)},
		{note(material.C, -1), note(material.B, 0)},
		{note(material.F, -1), note(material.E, 0)},
		{note(material.F, -2), note(material.E, -1)},
		{note(material.C, -2), note(material.B, -1)},
		{note(material.F, 2), note(material.G, 0)},
		{note(material.B, 2), note(material.C, 1)},
		{note(material.E, 2), note(material.F, 1)},
		{note(material.E, 1), note(material.E, 1)},
		{note(material.B, 1), note(material.B, 1)},
		{note(material.G, -1), note(material.G, -1)},
	}
	for _, tt := range tests {
		if got := notation.DisplaySpelling(tt.in); got != tt.want {
			t.Errorf("DisplaySpelling(%s) = %s, want %s", tt.in, got, tt.want)
		}
	}
	// Every note name keeps its pitch class; only the respelled ones change.
	for _, n := range allNotes() {
		got := notation.DisplaySpelling(n)
		respelled := n.Alter() == 2 || n.Alter() == -2 || n == note(material.C, -1) || n == note(material.F, -1)
		if got.PitchClass() != n.PitchClass() || (got != n) != respelled {
			t.Errorf("DisplaySpelling(%s) = %s; want the same pitch class, changed only if respelled (%t)", n, got, respelled)
		}
		if w := notation.FormatNoteName(got); strings.ContainsAny(w, "𝄪𝄫") || w == "C♭" || w == "F♭" {
			t.Errorf("DisplaySpelling(%s) = %s, still a spelling to respell", n, w)
		}
	}
}
