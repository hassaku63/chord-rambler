package reading_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// shapes covers the root alone, triads, sus, seventh chords, tensions, and
// basses that are members and that are not.
var shapes = []struct{ intervals, bass string }{
	{"", ""},
	{maj, ""}, {minTriad, ""}, {dim, ""}, {"M3 A5", ""}, {"P4 P5", ""}, {"M2 P5", ""},
	{dom7, ""}, {maj7, ""}, {min7, ""}, {half7, ""}, {"m3 d5 d7", ""},
	{"M3 P5 m7 m9 A11 m13", ""},
	{maj, "M3"}, {maj, "P5"}, {maj, "m7"}, {maj, "M2"}, {maj, "A4"},
}

// eachChord calls f with every major key on a tonic of at most one sharp or
// flat, and every chord of the shapes above on a root of at most a double
// sharp or flat. Only the failure the materials document is skipped: a chord
// beyond a double sharp or a double flat (material.ErrAlterOutOfRange).
func eachChord(t *testing.T, f func(key material.Scale, c material.Chord)) {
	t.Helper()
	for kl := material.C; kl <= material.B; kl++ {
		for ka := flat; ka <= sharp; ka++ {
			key := major(note(kl, ka))
			for l := material.C; l <= material.B; l++ {
				for a := -2; a <= 2; a++ {
					for _, s := range shapes {
						c, err := material.NewChord(note(l, a), ivs(s.intervals)...)
						if err == nil && s.bass != "" {
							c, err = c.WithBass(ivs(s.bass)[0])
						}
						if errors.Is(err, material.ErrAlterOutOfRange) {
							continue
						}
						if err != nil {
							t.Fatalf("chord on %s: %v", note(l, a), err)
						}
						f(key, c)
					}
				}
			}
		}
	}
}

// What a reading holds follows its kind: a function only for a diatonic
// chord, a target (2 to 6) only for a secondary dominant, modes only for a
// borrowed chord, and the key it was read in. Every kind is met at least once,
// so the checks are not vacuous.
func TestReadingFieldsFollowTheKind(t *testing.T) {
	seen := map[reading.Kind]bool{}
	eachChord(t, func(key material.Scale, c material.Chord) {
		got, err := reading.ReadChord(c, key)
		if err != nil {
			t.Errorf("ReadChord(%v, %s major): %v", c.Notes(), key.Tonic(), err)
			return
		}
		m := got.Main
		seen[m.Kind] = true
		if m.Key != key || len(got.Others) != 0 {
			t.Errorf("%v in %s major: Key = %s, Others = %v", c.Notes(), key.Tonic(), m.Key.Tonic(), got.Others)
		}
		if (m.Function != reading.NoFunction) != (m.Kind == reading.Diatonic) {
			t.Errorf("%v in %s major: %s with %s", c.Notes(), key.Tonic(), m.Kind, m.Function)
		}
		if (m.Target != 0) != (m.Kind == reading.SecondaryDominant) || m.Target < 0 || m.Target == 1 || m.Target > 6 {
			t.Errorf("%v in %s major: %s with target %d", c.Notes(), key.Tonic(), m.Kind, m.Target)
		}
		if (len(m.Modes) != 0) != (m.Kind == reading.Borrowed) {
			t.Errorf("%v in %s major: %s with modes %v", c.Notes(), key.Tonic(), m.Kind, m.Modes)
		}
		for _, mode := range m.Modes {
			if mode == material.Ionian {
				t.Errorf("%v in %s major: borrowed from Ionian", c.Notes(), key.Tonic())
			}
		}
	})
	for k := reading.FactsOnly; k <= reading.SecondaryDominant; k++ {
		if !seen[k] {
			t.Errorf("no chord was read as %s", k)
		}
	}
}

// A reading does not depend on where the key is: moving the tonic and the
// chord by the same interval keeps the kind, the function, the target and the
// modes. Pairs where the tonic or the chord would need more than a double
// sharp or flat are skipped; at least one pair must be compared.
func TestReadingMovesWithTheKey(t *testing.T) {
	up := ivs("M2")[0]
	compared := 0
	eachChord(t, func(key material.Scale, c material.Chord) {
		tonic, err := key.Tonic().Transpose(up)
		if errors.Is(err, material.ErrAlterOutOfRange) {
			return
		}
		if err != nil {
			t.Fatalf("Transpose(%s): %v", key.Tonic(), err)
		}
		movedKey, err := material.NewScale(tonic, material.Ionian)
		if errors.Is(err, material.ErrAlterOutOfRange) {
			return
		}
		if err != nil {
			t.Fatalf("NewScale(%s): %v", tonic, err)
		}
		movedChord, err := c.Transpose(up)
		if errors.Is(err, material.ErrAlterOutOfRange) {
			return
		}
		if err != nil {
			t.Fatalf("Transpose(%v): %v", c.Notes(), err)
		}
		before, err := reading.ReadChord(c, key)
		if err != nil {
			t.Fatalf("ReadChord: %v", err)
		}
		after, err := reading.ReadChord(movedChord, movedKey)
		if err != nil {
			t.Fatalf("ReadChord: %v", err)
		}
		compared++
		b, a := before.Main, after.Main
		if b.Kind != a.Kind || b.Function != a.Function || b.Target != a.Target || !slices.Equal(b.Modes, a.Modes) {
			t.Errorf("%v in %s major is %s %s %d %v; a major second up it is %s %s %d %v",
				c.Notes(), key.Tonic(), b.Kind, b.Function, b.Target, b.Modes, a.Kind, a.Function, a.Target, a.Modes)
		}
	})
	if compared == 0 {
		t.Error("no pair was compared")
	}
}

// The kind is judged by the notes that make the chord and a bass a minor
// seventh above the root (the owner's calls on #102), so adding a tension (9,
// 11 or 13), or a bass other than a minor seventh, keeps the reading. The second, fourth and sixth members are looked at, so they are
// not added here. Chords beyond a double sharp or
// flat are skipped; at least one must be compared.
func TestTensionsAndOtherBassesKeepTheReading(t *testing.T) {
	extras := []struct{ tension, bass string }{
		{"m9", ""}, {"A9", ""}, {"A11", ""}, {"m13", ""}, {"", "M2"}, {"", "A4"}, {"", "M6"},
	}
	compared := 0
	eachChord(t, func(key material.Scale, c material.Chord) {
		if _, ok := c.Bass(); ok || len(c.Tensions()) > 0 {
			return
		}
		before, err := reading.ReadChord(c, key)
		if err != nil {
			t.Fatalf("ReadChord: %v", err)
		}
		for _, x := range extras {
			more, err := material.NewChord(c.Root(), append(c.Intervals()[1:], ivs(x.tension)...)...)
			if err == nil && x.bass != "" {
				more, err = more.WithBass(ivs(x.bass)[0])
			}
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatalf("%v with %s%s: %v", c.Notes(), x.tension, x.bass, err)
			}
			after, err := reading.ReadChord(more, key)
			if err != nil {
				t.Fatalf("ReadChord: %v", err)
			}
			compared++
			b, a := before.Main, after.Main
			if b.Kind != a.Kind || b.Function != a.Function || b.Target != a.Target || !slices.Equal(b.Modes, a.Modes) {
				t.Errorf("%v in %s major is %s %s %d %v; with %s%s it is %s %s %d %v",
					c.Notes(), key.Tonic(), b.Kind, b.Function, b.Target, b.Modes,
					x.tension, x.bass, a.Kind, a.Function, a.Target, a.Modes)
			}
		}
	})
	if compared == 0 {
		t.Error("no chord was compared")
	}
}
