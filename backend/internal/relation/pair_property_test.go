package relation_test

import (
	"errors"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// allRoots returns every note name: seven letters, five alterations.
func allRoots() []material.NoteName {
	var out []material.NoteName
	for l := material.C; l <= material.B; l++ {
		for a := -2; a <= 2; a++ {
			out = append(out, note(l, a))
		}
	}
	return out
}

// chordsOn returns the chords of every shape on every root. Only the failure
// the materials document is skipped: a member or a bass beyond a double sharp
// or a double flat (material.ErrAlterOutOfRange).
func chordsOn(t *testing.T, roots []material.NoteName) []material.Chord {
	t.Helper()
	var out []material.Chord
	for _, root := range roots {
		for _, s := range shapes {
			c, err := material.NewChord(root, ivs(s.intervals)...)
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatalf("NewChord(%s, %s): %v", root, s.intervals, err)
			}
			if s.bass != "" {
				c, err = c.WithBass(ivs(s.bass)[0])
				if errors.Is(err, material.ErrAlterOutOfRange) {
					continue
				}
				if err != nil {
					t.Fatalf("WithBass(%s): %v", s.bass, err)
				}
			}
			out = append(out, c)
		}
	}
	return out
}

// Between any two roots, the root motion moves the first root to the second;
// when there is none, the error is ErrInvalidInterval.
func TestRootMotionMovesTheRoot(t *testing.T) {
	roots := allRoots()
	for _, from := range roots {
		for _, to := range roots {
			prev, next := chord(from, "", ""), chord(to, "", "")
			iv, err := relation.RootMotion(prev, next)
			if err != nil {
				if !errors.Is(err, material.ErrInvalidInterval) {
					t.Errorf("RootMotion(%s, %s): err = %v; want ErrInvalidInterval", from, to, err)
				}
				continue
			}
			if got, err := from.Transpose(iv); err != nil || got != to {
				t.Errorf("RootMotion(%s, %s) = %v, and %s moved by it is %s, %v", from, to, iv, from, got, err)
			}
		}
	}
}

// The common tones are in both chords, are the same either way round, and are
// what sounds when a chord meets itself.
func TestCommonTonesAreShared(t *testing.T) {
	var roots []material.NoteName
	for l := material.C; l <= material.B; l++ {
		for a := -1; a <= 1; a++ {
			roots = append(roots, note(l, a))
		}
	}
	chords := chordsOn(t, roots)
	for _, a := range chords {
		if got := relation.CommonTones(a, a); got != a.SoundingPitchClasses() {
			t.Errorf("CommonTones(%s %v, itself) = %v; want %v", a.Root(), a.Intervals(), got, a.SoundingPitchClasses())
		}
		for _, b := range chords {
			got := relation.CommonTones(a, b)
			if !got.IsSubsetOf(a.SoundingPitchClasses()) || !got.IsSubsetOf(b.SoundingPitchClasses()) {
				t.Errorf("CommonTones(%s %v, %s %v) = %v is not in both", a.Root(), a.Intervals(), b.Root(), b.Intervals(), got)
			}
			if back := relation.CommonTones(b, a); back != got {
				t.Errorf("CommonTones(%s %v, %s %v) = %v, the other way %v", a.Root(), a.Intervals(), b.Root(), b.Intervals(), got, back)
			}
		}
	}
}

// majorKeys returns the major scales on every tonic that can spell one.
func majorKeys(t *testing.T) []material.Scale {
	t.Helper()
	var keys []material.Scale
	for _, tonic := range allRoots() {
		key, err := material.NewScale(tonic, material.Ionian)
		if errors.Is(err, material.ErrAlterOutOfRange) {
			continue
		}
		if err != nil {
			t.Fatalf("NewScale(%s, Ionian): %v", tonic, err)
		}
		keys = append(keys, key)
	}
	return keys
}

// Both ways, in every major key and for every root and shape: a chord with
// the shape of a secondary dominant is the chord of V with that shape and the
// target, a major triad or a dominant seventh, and not in the key. And the
// chord of V/t (t from 2 to 6) with a major triad or a dominant seventh has
// the shape with target t exactly when it is not in the key (I7 is V7/IV, but
// I is in the key).
func TestSecondaryDominantShapeBothWays(t *testing.T) {
	five := func(s material.Shape, target int) material.Position {
		p, err := material.NewPosition(5, material.Natural, s)
		if err != nil {
			t.Fatal(err)
		}
		if p, err = p.WithTarget(target); err != nil {
			t.Fatalf("WithTarget(%d): %v", target, err)
		}
		return p
	}
	chords := chordsOn(t, allRoots())
	// Neither direction may pass by checking nothing.
	var found, built int
	for _, key := range majorKeys(t) {
		tonic := key.Tonic()
		for _, c := range chords {
			target, ok, err := relation.SecondaryDominantShape(c, key)
			if err != nil {
				t.Fatalf("%s in %s major: %v", c.Root(), tonic, err)
			}
			if !ok {
				continue
			}
			found++
			back, err := five(material.ShapeOf(c), target).Chord(tonic)
			if err != nil || !chordsEqual(back, c) {
				t.Errorf("%s %v in %s major: target %d, but V/%d is %v, %v", c.Root(), c.Intervals(), tonic, target, target, back, err)
			}
			if !dominantKind(c) || relation.InScale(c, key) {
				t.Errorf("%s %v in %s major: target %d, but the kind or the key does not fit", c.Root(), c.Intervals(), tonic, target)
			}
		}
		for target := 2; target <= 6; target++ {
			for _, s := range shapes {
				p := five(material.ShapeOf(chord(material.NoteName{}, s.intervals, s.bass)), target)
				c, err := p.Chord(tonic)
				if errors.Is(err, material.ErrAlterOutOfRange) {
					continue
				}
				if err != nil {
					t.Fatalf("%s in %s major: Chord: %v", describe(p), tonic, err)
				}
				if !dominantKind(c) {
					continue
				}
				built++
				want := !relation.InScale(c, key)
				if got, ok, err := relation.SecondaryDominantShape(c, key); err != nil || ok != want || (ok && got != target) {
					t.Errorf("V/%d %v in %s major (%s): SecondaryDominantShape = %d, %v, %v; want %d, %v",
						target, c.Intervals(), tonic, c.Root(), got, ok, err, target, want)
				}
			}
		}
	}
	if found == 0 || built == 0 {
		t.Errorf("found %d chords with the shape and built %d chords of V/t; want some of each", found, built)
	}
}

// dominantKind reports a major triad or a dominant seventh, whatever the
// tensions.
func dominantKind(c material.Chord) bool {
	s := c.Seventh()
	return c.Triad() == material.TriadMajor && (s == material.NoSeventh || s == material.SeventhMinor)
}
