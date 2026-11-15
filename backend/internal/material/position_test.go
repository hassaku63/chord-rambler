package material_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// withBassOf returns the chord with the bass, an interval such as "M3", above
// the root.
func withBassOf(c material.Chord, bass string) material.Chord {
	c, err := c.WithBass(ivs(bass)[0])
	if err != nil {
		panic(err)
	}
	return c
}

// position builds a position for a test table. A target of 0 means none.
func position(degree int, alter material.Alteration, intervals string, target int) material.Position {
	p, err := material.NewPosition(degree, alter, material.ShapeOf(chord(material.NoteName{}, intervals)))
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

// describeChord shows a chord in a failure message: the root and the
// intervals.
func describeChord(c material.Chord) string {
	return fmt.Sprintf("%s %v", c.Root(), c.Intervals())
}

func TestNewPositionRejects(t *testing.T) {
	var shape material.Shape
	tests := []struct {
		degree int
		alter  material.Alteration
		want   error
	}{
		{0, material.Natural, material.ErrDegree},
		{8, material.Natural, material.ErrDegree},
		{1, material.Alteration(2), material.ErrAlteration},
		{1, material.Alteration(-2), material.ErrAlteration},
	}
	for _, tt := range tests {
		if _, err := material.NewPosition(tt.degree, tt.alter, shape); !errors.Is(err, tt.want) {
			t.Errorf("NewPosition(%d, %d) = %v; want %v", tt.degree, tt.alter, err, tt.want)
		}
	}
	p, err := material.NewPosition(5, material.Natural, shape)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []int{0, 1, 7, 8} {
		if _, err := p.WithTarget(target); !errors.Is(err, material.ErrTarget) {
			t.Errorf("WithTarget(%d) = %v; want %v", target, err, material.ErrTarget)
		}
	}
}

// The bass of the shape comes with the chord: V with the third in the bass is
// G/B in C. A root, a member or the root of the target beyond a double sharp
// fails. In G♯ major the seventh degree is F𝄪, so ♯VII is beyond it. The
// augmented fifth of C𝄪 would be G with three sharps. The fifth degree of B𝄪
// major would be F with three sharps.
func TestPositionChord(t *testing.T) {
	withBass, err := material.NewPosition(5, material.Natural, material.ShapeOf(withBassOf(chord(material.NoteName{}, "M3 P5"), "M3")))
	if err != nil {
		t.Fatal(err)
	}
	c, err := withBass.Chord(note(material.C, 0))
	if bass, ok := c.Bass(); err != nil || c.Root() != note(material.G, 0) || !ok || bass != interval(3, material.Major) ||
		!slices.Equal(c.Intervals(), ivs("P1 M3 P5")) {
		t.Errorf("V with the third in the bass in C = %v, %v; want G with M3 P5 over the bass M3", c, err)
	}
	tests := []struct {
		name  string
		tonic material.NoteName
		p     material.Position
	}{
		{"♯VII in G♯ (beyond F𝄪)", note(material.G, 1), position(7, material.Sharp, "M3 P5", 0)},
		{"I+ in C𝄪 (the fifth)", note(material.C, 2), position(1, material.Natural, "M3 A5", 0)},
		{"V/V in B𝄪 (the target)", note(material.B, 2), position(5, material.Natural, "M3 P5", 5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.p.Chord(tt.tonic); !errors.Is(err, material.ErrAlterOutOfRange) {
				t.Errorf("Chord(%s) = %s, %v; want error %v", tt.tonic, describeChord(got), err, material.ErrAlterOutOfRange)
			}
		})
	}
}
