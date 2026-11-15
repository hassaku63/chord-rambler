package relation_test

import (
	"errors"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

func pcs(ns ...int) material.PitchClassSet {
	var out []material.PitchClass
	for _, n := range ns {
		pc, err := material.NewPitchClass(n)
		if err != nil {
			panic(err)
		}
		out = append(out, pc)
	}
	return material.NewPitchClassSet(out...)
}

// The root motion is the simple interval up from one root to the next, named
// by letters (material.IntervalBetween; the first four cases are those of
// #87). The bass is not looked at: C/E to F moves from C.
func TestRootMotion(t *testing.T) {
	tests := []struct {
		name       string
		prev, next material.Chord
		want       string
	}{
		{"G → C", chord(note(material.G, nat), "M3 P5", ""), chord(note(material.C, nat), "M3 P5", ""), "P4"},
		{"C → A", chord(note(material.C, nat), "M3 P5", ""), chord(note(material.A, nat), "m3 P5", ""), "M6"},
		{"D → G", chord(note(material.D, nat), "m3 P5 m7", ""), chord(note(material.G, nat), "M3 P5 m7", ""), "P4"},
		{"B♭ → C", chord(note(material.B, flat), "M3 P5", ""), chord(note(material.C, nat), "M3 P5", ""), "M2"},
		{"C → C (the same root)", chord(note(material.C, nat), "M3 P5", ""), chord(note(material.C, nat), "M3 P5 m7", ""), "P1"},
		{"C♯ → C (the same letter, down)", chord(note(material.C, sharp), "m3 P5", ""), chord(note(material.C, nat), "M3 P5", ""), "d1"},
		{"C/E → F (the bass is not looked at)", chord(note(material.C, nat), "M3 P5", "M3"), chord(note(material.F, nat), "M3 P5", ""), "P4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := ivs(tt.want)[0]
			if got, err := relation.RootMotion(tt.prev, tt.next); err != nil || got != want {
				t.Errorf("RootMotion = %v, %v; want %v", got, err, want)
			}
		})
	}
	// C♯ to E𝄫 is a third of one half step, which has no quality.
	c, e := chord(note(material.C, sharp), "", ""), chord(note(material.E, -2), "", "")
	if got, err := relation.RootMotion(c, e); !errors.Is(err, material.ErrInvalidInterval) {
		t.Errorf("RootMotion(C♯, E𝄫) = %v, %v; want ErrInvalidInterval", got, err)
	}
}

// Common tones are the pitch classes that sound in both chords: C (C E G) and
// Am (A C E) share C and E; G7 (G B D F) and C share G; C and F♯ (F♯ A♯ C♯)
// share none (#87). A bass that is not a member sounds: C/D (C E G D) and G
// (G B D) share D and G.
func TestCommonTones(t *testing.T) {
	tests := []struct {
		name string
		a, b material.Chord
		want material.PitchClassSet
	}{
		{"C and Am", chord(note(material.C, nat), "M3 P5", ""), chord(note(material.A, nat), "m3 P5", ""), pcs(0, 4)},
		{"G7 and C", chord(note(material.G, nat), "M3 P5 m7", ""), chord(note(material.C, nat), "M3 P5", ""), pcs(7)},
		{"C and F♯", chord(note(material.C, nat), "M3 P5", ""), chord(note(material.F, sharp), "M3 P5", ""), pcs()},
		{"C/D and G", chord(note(material.C, nat), "M3 P5", "M2"), chord(note(material.G, nat), "M3 P5", ""), pcs(2, 7)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relation.CommonTones(tt.a, tt.b); got != tt.want {
				t.Errorf("CommonTones = %v; want %v", got, tt.want)
			}
		})
	}
}

// The shape of a secondary dominant in C major (docs/architecture/
// theory-conventions.md, 2; the cases are those of #87). Hutchinson 17.4
// Analyzing Secondary Dominants
// (https://musictheory.pugetsound.edu/mt21c/AnalyzingSecondaryDominants.html):
// a major triad or a major–minor seventh chord whose root is a perfect fifth
// above the root of a diatonic chord. 17.2 Tonicization
// (https://musictheory.pugetsound.edu/mt21c/Tonicization.html) names D7 in C
// major V7/V. Target 0 means the chord has not the shape.
func TestSecondaryDominantShapeInCMajor(t *testing.T) {
	seventh, triad := "M3 P5 m7", "M3 P5"
	tests := []struct {
		name   string
		chord  material.Chord
		target int
	}{
		{"D7 is V/V", chord(note(material.D, nat), seventh, ""), 5},
		{"A7 is V/ii", chord(note(material.A, nat), seventh, ""), 2},
		{"E is V/vi", chord(note(material.E, nat), triad, ""), 6},
		{"B7 is V/iii", chord(note(material.B, nat), seventh, ""), 3},
		{"C7 is V/IV", chord(note(material.C, nat), seventh, ""), 4},
		{"G7 has no note outside the key", chord(note(material.G, nat), seventh, ""), 0},
		{"F♯7: B is the root of vii°, a diminished triad", chord(note(material.F, sharp), seventh, ""), 0},
		{"B♭7: E♭ is not a root of the key", chord(note(material.B, flat), seventh, ""), 0},
		// The tensions are not asked.
		{"D7(♭9) is V/V", chord(note(material.D, nat), seventh+" m9", ""), 5},
		{"D is V/V", chord(note(material.D, nat), triad, ""), 5},
		{"Dmaj7 has a major seventh", chord(note(material.D, nat), "M3 P5 M7", ""), 0},
		{"D+ is augmented", chord(note(material.D, nat), "M3 A5", ""), 0},
		{"Dm is minor", chord(note(material.D, nat), "m3 P5", ""), 0},
		{"C is I, in the key", chord(note(material.C, nat), triad, ""), 0},
		// The fifth below is judged by spelling: C♭ sounds like B, but F♭ is
		// not E.
		{"C♭7: F♭ is not a root of the key", chord(note(material.C, flat), seventh, ""), 0},
		// What sounds is compared with the key, so a bass or a tension outside the
		// key counts.
		{"C/B♭ is V/IV", chord(note(material.C, nat), triad, "m7"), 4},
		{"C(♯11): only the tension F♯ is outside the key", chord(note(material.C, nat), triad+" A11", ""), 4},
		{"G/C♯: C is I", chord(note(material.G, nat), triad, "A4"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := relation.SecondaryDominantShape(tt.chord, cMajor)
			if err != nil || got != tt.target || ok != (tt.target != 0) {
				t.Errorf("SecondaryDominantShape = %d, %v, %v; want %d, %v", got, ok, err, tt.target, tt.target != 0)
			}
		})
	}
}

// The degree is counted in the major scale, so a key in another mode is
// refused, as PositionOf refuses it.
func TestSecondaryDominantShapeOtherModes(t *testing.T) {
	d7 := chord(note(material.D, nat), "M3 P5 m7", "")
	for m := material.Dorian; m <= material.Locrian; m++ {
		key, err := material.NewScale(note(material.C, nat), m)
		if err != nil {
			t.Fatal(err)
		}
		if got, ok, err := relation.SecondaryDominantShape(d7, key); !errors.Is(err, relation.ErrNotMajor) {
			t.Errorf("C %s: SecondaryDominantShape(D7) = %d, %v, %v; want ErrNotMajor", m, got, ok, err)
		}
	}
}
