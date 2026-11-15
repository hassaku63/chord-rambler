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
	// The test may not pass by every pair failing.
	moved := 0
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
			moved++
			if got, err := from.Transpose(iv); err != nil || got != to {
				t.Errorf("RootMotion(%s, %s) = %v, and %s moved by it is %s, %v", from, to, iv, from, got, err)
			}
		}
	}
	if moved == 0 {
		t.Error("RootMotion failed for every pair of roots; want some intervals")
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

// five returns the position of V with the shape and the target.
func five(t *testing.T, s material.Shape, target int) material.Position {
	t.Helper()
	p, err := material.NewPosition(5, material.Natural, s)
	if err != nil {
		t.Fatal(err)
	}
	if p, err = p.WithTarget(target); err != nil {
		t.Fatalf("WithTarget(%d): %v", target, err)
	}
	return p
}

// Both ways, in every major key and for every root and shape: a chord with
// the shape of a secondary dominant is the chord of V with that shape and the
// target, a major triad or a dominant seventh, and not in the key. And the
// chord of V/t (t from 2 to 6) with a major triad or a dominant seventh has
// the shape with target t exactly when the notes it is judged by leave the
// key. The cases of the kind and of the key, one by one, are pinned by the
// table tests (TestSecondaryDominantShapeInCMajor); this test checks that the
// shape and the target agree with Position in every key.
func TestSecondaryDominantShapeBothWays(t *testing.T) {
	minorSeventh := ivs("m7")[0]
	chords := chordsOn(t, allRoots())
	// Neither direction may pass by checking nothing, and the chords of V/t
	// must have both answers.
	var found, built, builtOutside int
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
			back, err := five(t, material.ShapeOf(c), target).Chord(tonic)
			if err != nil || !chordsEqual(back, c) {
				t.Errorf("%s %v in %s major: target %d, but V/%d is %v, %v", c.Root(), c.Intervals(), tonic, target, target, back, err)
			}
			if !dominantKind(c) || relation.InScale(c, key) {
				t.Errorf("%s %v in %s major: target %d, but the kind or the key does not fit", c.Root(), c.Intervals(), tonic, target)
			}
		}
		for target := 2; target <= 6; target++ {
			for _, s := range shapes {
				p := five(t, material.ShapeOf(chord(material.NoteName{}, s.intervals, s.bass)), target)
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
				// The major third of V/t is the leading tone of t, outside the
				// major key unless t is 4: there V/IV is I, in the key, and only
				// a minor seventh, as a member or as the bass, leaves it.
				bass, hasBass := c.Bass()
				want := target != 4 || c.Seventh() == material.SeventhMinor || (hasBass && bass == minorSeventh)
				if want {
					builtOutside++
				}
				if got, ok, err := relation.SecondaryDominantShape(c, key); err != nil || ok != want || (ok && got != target) {
					t.Errorf("V/%d %v in %s major (%s): SecondaryDominantShape = %d, %v, %v; want %d, %v",
						target, c.Intervals(), tonic, c.Root(), got, ok, err, target, want)
				}
			}
		}
	}
	if found == 0 || builtOutside == 0 || builtOutside == built {
		t.Errorf("found %d chords with the shape and built %d chords of V/t, %d of them outside the key; want some of each, and some in the key", found, built, builtOutside)
	}
}

// Adding a tension, or a bass other than a minor seventh above the root, to a
// major triad or a dominant seventh does not change the shape: in every major
// key, for the chord of V/t (t from 2 to 6) as a triad, a dominant seventh
// and a triad over the minor seventh. The decorations include notes outside
// the key on a chord in the key (C(♯11), C/F♯ and C/A♭ in C major).
func TestSecondaryDominantShapeIgnoresDecorations(t *testing.T) {
	bases := []struct{ intervals, bass string }{{"M3 P5", ""}, {"M3 P5 m7", ""}, {"M3 P5", "m7"}}
	tensions := ivs("m9 M9 A9 P11 A11 m13 M13")
	basses := ivs("P1 m2 M2 A2 m3 M3 P4 A4 d5 P5 A5 m6 M6 A6 d7 M7")
	// The test may not pass by checking nothing, nor without the case it is
	// for: a note outside the key added to a chord in the key.
	var shaped, outsideOnlyByDecoration int
	for _, key := range majorKeys(t) {
		tonic := key.Tonic()
		for target := 2; target <= 6; target++ {
			for _, b := range bases {
				p := five(t, material.ShapeOf(chord(material.NoteName{}, b.intervals, b.bass)), target)
				base, err := p.Chord(tonic)
				if errors.Is(err, material.ErrAlterOutOfRange) {
					continue
				}
				if err != nil {
					t.Fatalf("%s in %s major: Chord: %v", describe(p), tonic, err)
				}
				wantTarget, wantOK, err := relation.SecondaryDominantShape(base, key)
				if err != nil {
					t.Fatalf("V/%d %v in %s major: %v", target, base.Intervals(), tonic, err)
				}
				if wantOK {
					shaped++
				}
				for _, c := range decorated(t, base, tensions, basses) {
					if relation.InScale(base, key) && !relation.InScale(c, key) {
						outsideOnlyByDecoration++
					}
					got, ok, err := relation.SecondaryDominantShape(c, key)
					if err != nil || got != wantTarget || ok != wantOK {
						t.Errorf("%s %v (bass %v) in %s major: SecondaryDominantShape = %d, %v, %v; want %d, %v as without the decoration",
							c.Root(), c.Intervals(), bassOf(c), tonic, got, ok, err, wantTarget, wantOK)
					}
				}
			}
		}
	}
	if shaped == 0 || outsideOnlyByDecoration == 0 {
		t.Errorf("%d bases with the shape and %d decorated chords outside the key only by the decoration; want some of each", shaped, outsideOnlyByDecoration)
	}
}

// decorated returns the chord with each tension added, and, when it has no
// bass, with each bass. Only the failure the materials document is skipped:
// a note beyond a double sharp or a double flat (material.ErrAlterOutOfRange).
func decorated(t *testing.T, base material.Chord, tensions, basses []material.Interval) []material.Chord {
	t.Helper()
	var out []material.Chord
	bass, hasBass := base.Bass()
	for _, tension := range tensions {
		c, err := material.NewChord(base.Root(), append(base.Intervals()[1:], tension)...)
		if err == nil && hasBass {
			c, err = c.WithBass(bass)
		}
		if errors.Is(err, material.ErrAlterOutOfRange) {
			continue
		}
		if err != nil {
			t.Fatalf("%s %v with %v: %v", base.Root(), base.Intervals(), tension, err)
		}
		out = append(out, c)
	}
	if hasBass {
		return out
	}
	for _, b := range basses {
		c, err := base.WithBass(b)
		if errors.Is(err, material.ErrAlterOutOfRange) {
			continue
		}
		if err != nil {
			t.Fatalf("%s %v over %v: %v", base.Root(), base.Intervals(), b, err)
		}
		out = append(out, c)
	}
	return out
}

func bassOf(c material.Chord) string {
	if b, ok := c.Bass(); ok {
		return b.String()
	}
	return "none"
}

// dominantKind reports a major triad or a dominant seventh, whatever the
// tensions.
func dominantKind(c material.Chord) bool {
	s := c.Seventh()
	return c.Triad() == material.TriadMajor && (s == material.NoSeventh || s == material.SeventhMinor)
}
