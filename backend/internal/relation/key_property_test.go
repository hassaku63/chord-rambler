package relation_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// shapes covers the root alone, triads, sus, seventh chords, tensions, and
// basses that are members and that are not.
var shapes = []struct{ intervals, bass string }{
	{"", ""},
	{"M3 P5", ""}, {"m3 P5", ""}, {"m3 d5", ""}, {"M3 A5", ""}, {"P4 P5", ""}, {"M2 P5", ""},
	{"M3 P5 m7", ""}, {"M3 P5 M7", ""}, {"m3 P5 m7", ""}, {"m3 d5 m7", ""}, {"m3 d5 d7", ""},
	{"M3 P5 m7 m9 A11 m13", ""},
	{"M3 P5", "M3"}, {"M3 P5", "P5"}, {"M3 P5", "m7"}, {"M3 P5", "M2"},
}

// chordsInMajorKeys calls f with every position without a target on every
// tonic of a major scale, every degree, alteration and shape above, and the
// chord of the position. Only the failure the materials document is skipped:
// a scale or a chord beyond a double sharp or a double flat
// (material.ErrAlterOutOfRange). In C major no natural degree is skipped: the
// root is a natural note, and a natural note moved by any interval of the five
// qualities needs at most a double sharp or a double flat.
func chordsInMajorKeys(t *testing.T, f func(key material.Scale, p material.Position, c material.Chord)) {
	t.Helper()
	for l := material.C; l <= material.B; l++ {
		for a := -2; a <= 2; a++ {
			tonic := note(l, a)
			key, err := material.NewScale(tonic, material.Ionian)
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatalf("NewScale(%s, Ionian): %v", tonic, err)
			}
			for degree := 1; degree <= 7; degree++ {
				for alter := material.Flat; alter <= material.Sharp; alter++ {
					for _, s := range shapes {
						p := position(degree, alter, s.intervals, s.bass)
						c, err := p.Chord(tonic)
						if errors.Is(err, material.ErrAlterOutOfRange) && (tonic != material.NoteName{} || alter != material.Natural) {
							continue
						}
						if err != nil {
							t.Errorf("%s in %s major: Chord: %v", describe(p), tonic, err)
							continue
						}
						f(key, p, c)
					}
				}
			}
		}
	}
}

// The position of the chord of a position is the position.
func TestPositionOfRoundTrip(t *testing.T) {
	chordsInMajorKeys(t, func(key material.Scale, p material.Position, c material.Chord) {
		if got, err := relation.PositionOf(c, key); err != nil || got != p {
			t.Errorf("%s major: PositionOf(Chord(%s)) = %s, %v", key.Tonic(), describe(p), describe(got), err)
		}
	})
}

// The parallel modes agree with the scales of the modes on the tonic, wherever
// such a scale can be spelled.
func TestParallelModesAgreeWithScales(t *testing.T) {
	chordsInMajorKeys(t, func(key material.Scale, _ material.Position, c material.Chord) {
		got := relation.ParallelModes(c, key)
		for m := material.Ionian; m <= material.Locrian; m++ {
			s, err := material.NewScale(key.Tonic(), m)
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatalf("NewScale(%s, %s): %v", key.Tonic(), m, err)
			}
			if want := s.Contains(c.SoundingPitchClasses()); slices.Contains(got, m) != want {
				t.Errorf("%s with %v: %s in ParallelModes = %v; the scale contains it: %v",
					c.Root(), c.Intervals(), m, !want, want)
			}
		}
	})
}
