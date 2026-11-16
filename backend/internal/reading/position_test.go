package reading_test

import (
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

func TestRomanPositionInCMajor(t *testing.T) {
	key := major(note(material.C, nat))
	// want is the position written by hand: degree, alteration, the shape's
	// intervals above the root (no bass) and the target (0 for none).
	tests := []struct {
		name      string
		chord     material.Chord
		none      bool
		degree    int
		alter     material.Alteration
		intervals string
		target    int
	}{
		{"C7 is V7/IV", chord(note(material.C, nat), "M3 P5 m7", ""), false, 5, material.Natural, "M3 P5 m7", 4},
		{"A7/C# is V7/ii", chord(note(material.A, nat), "M3 P5 m7", "M3"), false, 5, material.Natural, "M3 P5 m7", 2},
		{"D7(9) is V7(9)/V", chord(note(material.D, nat), "M3 P5 m7 M9", ""), false, 5, material.Natural, "M3 P5 m7 M9", 5},
		{"G/B is V", chord(note(material.G, nat), "M3 P5", "M3"), false, 5, material.Natural, "M3 P5", 0},
		{"Bb is bVII", chord(note(material.B, flat), "M3 P5", ""), false, 7, material.Flat, "M3 P5", 0},
		{"Bdim7 is vii dim7", chord(note(material.B, nat), "m3 d5 d7", ""), false, 7, material.Natural, "m3 d5 d7", 0},
		{"Csus4/G is Isus4", chord(note(material.C, nat), "P4 P5", "P5"), false, 1, material.Natural, "P4 P5", 0},
		{"Dbb has no position", chord(note(material.D, -2), "M3 P5", ""), true, 0, 0, "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := reading.ReadChord(tt.chord, key)
			if err != nil {
				t.Fatal(err)
			}
			got, ok, err := reading.RomanPosition(tt.chord, key, r.Main)
			if err != nil {
				t.Fatal(err)
			}
			if ok == tt.none {
				t.Fatalf("ok = %v, want %v", ok, !tt.none)
			}
			if tt.none {
				return
			}
			if got.Degree() != tt.degree || got.Alteration() != tt.alter {
				t.Errorf("position = degree %d alteration %d, want %d and %d",
					got.Degree(), got.Alteration(), tt.degree, tt.alter)
			}
			want := chord(note(material.C, nat), tt.intervals, "")
			if !equalChords(t, got.Shape(), want) {
				t.Errorf("shape = %v, want intervals %s", got.Shape().Intervals(), tt.intervals)
			}
			if _, hasBass := got.Shape().Bass(); hasBass {
				t.Error("the position has a bass")
			}
			target, hasTarget := got.Target()
			if target != tt.target || hasTarget != (tt.target != 0) {
				t.Errorf("target = %d (%v), want %d", target, hasTarget, tt.target)
			}
		})
	}
}

// equalChords reports whether the shape, put on C, is the chord on C.
func equalChords(t *testing.T, s material.Shape, want material.Chord) bool {
	t.Helper()
	got, err := s.On(note(material.C, nat))
	if err != nil {
		t.Fatal(err)
	}
	gn, wn := got.Notes(), want.Notes()
	if len(gn) != len(wn) {
		return false
	}
	for i := range gn {
		if gn[i] != wn[i] {
			return false
		}
	}
	return true
}

// The position, put on the tonic, is the chord without its bass: the same
// notes, for every chord that has a position.
func TestRomanPositionSpellsTheChordWithoutItsBass(t *testing.T) {
	positioned := 0
	eachChord(t, func(key material.Scale, c material.Chord) {
		r, err := reading.ReadChord(c, key)
		if err != nil {
			t.Fatal(err)
		}
		pos, ok, err := reading.RomanPosition(c, key, r.Main)
		if err != nil {
			t.Fatalf("%v in %v: %v", c.Notes(), key.Tonic(), err)
		}
		if !ok {
			return
		}
		positioned++
		back, err := pos.Chord(key.Tonic())
		if err != nil {
			t.Fatalf("%v in %v: %v", c.Notes(), key.Tonic(), err)
		}
		if _, hasBass := back.Bass(); hasBass {
			t.Errorf("%v in %v: the position has a bass", c.Notes(), key.Tonic())
		}
		want, err := material.NewChord(c.Root(), c.Intervals()[1:]...)
		if err != nil {
			t.Fatal(err)
		}
		gn, wn := back.Notes(), want.Notes()
		if len(gn) != len(wn) {
			t.Fatalf("%v in %v: notes %v, want %v", c.Notes(), key.Tonic(), gn, wn)
		}
		for i := range gn {
			if gn[i] != wn[i] {
				t.Errorf("%v in %v: notes %v, want %v", c.Notes(), key.Tonic(), gn, wn)
				break
			}
		}
	})
	if positioned == 0 {
		t.Fatal("no chord had a position, so the check is vacuous")
	}
}
