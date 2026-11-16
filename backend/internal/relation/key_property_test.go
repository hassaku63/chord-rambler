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

// Every note name (35: seven letters, five alterations) as a
// root, with every shape, in every key: if the chord has a position, the chord
// of the position is the chord; if it has none, the error is ErrNoPosition.
func TestPositionOfBothWays(t *testing.T) {
	for kl := material.C; kl <= material.B; kl++ {
		for ka := -1; ka <= 1; ka++ {
			tonic := note(kl, ka)
			key := major(tonic)
			for l := material.C; l <= material.B; l++ {
				for a := -2; a <= 2; a++ {
					for _, s := range shapes {
						c, err := material.NewChord(note(l, a), ivs(s.intervals)...)
						if errors.Is(err, material.ErrAlterOutOfRange) {
							continue
						}
						if err != nil {
							t.Fatalf("NewChord: %v", err)
						}
						if s.bass != "" {
							if c, err = c.WithBass(ivs(s.bass)[0]); errors.Is(err, material.ErrAlterOutOfRange) {
								continue
							} else if err != nil {
								t.Fatalf("WithBass: %v", err)
							}
						}
						p, err := relation.PositionOf(c, key)
						if err != nil {
							if !errors.Is(err, relation.ErrNoPosition) {
								t.Errorf("%s in %s major: err = %v; want ErrNoPosition", c.Root(), tonic, err)
							}
							continue
						}
						back, err := p.Chord(tonic)
						if err != nil || !chordsEqual(back, c) {
							t.Errorf("%s in %s major: Chord(PositionOf) = %v, %v; want the chord", c.Root(), tonic, back, err)
						}
					}
				}
			}
		}
	}
}

// chordsEqual compares the root, the intervals and the bass.
func chordsEqual(a, b material.Chord) bool {
	ab, aok := a.Bass()
	bb, bok := b.Bass()
	return a.Root() == b.Root() && slices.Equal(a.Intervals(), b.Intervals()) && aok == bok && ab == bb
}
