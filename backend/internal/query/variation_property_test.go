package query_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// The twelve major keys (docs/architecture/engine.md, 土台).
var tonics = []string{"C", "G", "D", "A", "E", "B", "F♯", "D♭", "A♭", "E♭", "B♭", "F"}

// asked holds chords of the three qualities, with sevenths, tensions and
// basses, as intervals above the root.
var asked = []struct{ intervals, bass string }{
	{"M3 P5", ""}, {"m3 P5", ""}, {"m3 d5", ""},
	{"M3 P5 m7", ""}, {"M3 P5 M7", ""}, {"m3 P5 m7", ""}, {"m3 d5 m7", ""}, {"m3 d5 d7", ""},
	{"M3 P5 M6", ""}, {"M3 P5 M9", ""}, {"M3 P5 m7 m9", ""}, {"m3 P5 m7 M9 P11", ""},
	{"M3 P5", "M3"}, {"M3 P5", "m7"}, {"m3 P5", "P5"},
}

// eachAsked calls f with every major key of the twelve and every chord of the
// shapes above on a root of at most one sharp or flat, skipping only a chord
// or a variation beyond a double sharp or a double flat.
func eachAsked(t *testing.T, f func(key material.Scale, c material.Chord, vs []query.Variation)) {
	t.Helper()
	for _, tonic := range tonics {
		key := major(note(tonic))
		for l := material.C; l <= material.B; l++ {
			for a := -1; a <= 1; a++ {
				root, err := material.NewNoteName(l, a)
				if err != nil {
					t.Fatal(err)
				}
				for _, s := range asked {
					c, err := material.NewChord(root, ivs(s.intervals)...)
					if err == nil && s.bass != "" {
						c, err = c.WithBass(ivs(s.bass)[0])
					}
					if errors.Is(err, material.ErrAlterOutOfRange) {
						continue
					}
					if err != nil {
						t.Fatalf("chord on %s: %v", root, err)
					}
					vs, err := query.Variations(c, key)
					if errors.Is(err, material.ErrAlterOutOfRange) {
						continue
					}
					if err != nil {
						t.Fatalf("Variations(%v, %s major): %v", c.Notes(), tonic, err)
					}
					f(key, c, vs)
				}
			}
		}
	}
}

// The labels are the facts of reading the variation itself (誤り 5 in
// docs/architecture/explanations.md): the main reading of reading.ReadChord,
// the cautions of reading.TensionCautions, the members of the chord, and the
// members whose spelling is not in the scale of the key.
func TestLabelsAreTheReadingOfTheVariation(t *testing.T) {
	eachAsked(t, func(key material.Scale, c material.Chord, vs []query.Variation) {
		for _, v := range vs {
			r, err := reading.ReadChord(v.Chord, key)
			if err != nil {
				t.Fatal(err)
			}
			m, g := r.Main, v.Reading.Main
			if g.Key != m.Key || g.Kind != m.Kind || g.Function != m.Function || g.Target != m.Target || !slices.Equal(g.Modes, m.Modes) {
				t.Errorf("%v in %s major: reading label %+v; ReadChord %+v", v.Chord.Notes(), key.Tonic(), g, m)
			}
			cautions, err := reading.TensionCautions(v.Chord, key)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(v.Cautions.Cautions, cautions) {
				t.Errorf("%v in %s major: cautions label %v; TensionCautions %v", v.Chord.Notes(), key.Tonic(), v.Cautions.Cautions, cautions)
			}
			if !slices.Equal(v.Members.Notes, v.Chord.Notes()) {
				t.Errorf("%v: members label %v", v.Chord.Notes(), v.Members.Notes)
			}
			var outOfKey []material.NoteName
			for _, n := range v.Chord.Notes() {
				if !slices.Contains(key.Notes(), n) {
					outOfKey = append(outOfKey, n)
				}
			}
			if !slices.Equal(v.OutOfKey.Notes, outOfKey) {
				t.Errorf("%v in %s major (asked %v): out-of-key label %v; want %v", v.Chord.Notes(), key.Tonic(), c.Notes(), v.OutOfKey.Notes, outOfKey)
			}
		}
	})
}

// A variation keeps the root, has no bass, and keeps major and minor
// (docs/glossary.md, 変形): its triad is the triad of the chord asked about or
// a sus. At most one is the chord asked about, and only when that has no bass.
func TestVariationsKeepTheRootAndTheQuality(t *testing.T) {
	eachAsked(t, func(key material.Scale, c material.Chord, vs []query.Variation) {
		inputs := 0
		for _, v := range vs {
			if v.Chord.Root() != c.Root() {
				t.Errorf("%v: variation on %s", c.Notes(), v.Chord.Root())
			}
			if _, ok := v.Chord.Bass(); ok {
				t.Errorf("%v: variation %v has a bass", c.Notes(), v.Chord.Notes())
			}
			if tr := v.Chord.Triad(); tr != c.Triad() && tr != material.TriadSus2 && tr != material.TriadSus4 {
				t.Errorf("%v (%s): variation %v is %s", c.Notes(), c.Triad(), v.Chord.Notes(), tr)
			}
			if v.IsInput {
				inputs++
			}
		}
		_, hasBass := c.Bass()
		if inputs > 1 || (hasBass && inputs != 0) {
			t.Errorf("%v: %d variations marked as the chord asked about", c.Notes(), inputs)
		}
	})
}

// Moving the chord and the key by the same interval moves the list: in each
// of the twelve major keys, the chords of TestVariationsInCMajor give the
// same shapes, marks and labels as in C, and add notes from outside the key
// at the same intervals above the root.
func TestVariationsFollowTheKey(t *testing.T) {
	inC := major(note("C"))
	for _, s := range []struct{ root, intervals, bass string }{
		{"G", "M3 P5", ""}, {"G", "M3 P5 m7", ""}, {"A", "m3 P5", ""}, {"B", "m3 d5", ""},
		{"F", "m3 P5", ""}, {"D", "M3 P5 m7", ""}, {"C", "M3 P5", "M3"},
	} {
		c := chord(s.root, s.intervals, s.bass)
		base, err := query.Variations(c, inC)
		if err != nil {
			t.Fatal(err)
		}
		for _, tonic := range tonics {
			up, err := material.IntervalBetween(note("C"), note(tonic))
			if err != nil {
				t.Fatal(err)
			}
			moved, err := c.Transpose(up)
			if err != nil {
				t.Fatal(err)
			}
			got, err := query.Variations(moved, major(note(tonic)))
			if err != nil {
				t.Fatalf("%v in %s major: %v", moved.Notes(), tonic, err)
			}
			if len(got) != len(base) {
				t.Fatalf("%v in %s major: %d variations; %d in C", moved.Notes(), tonic, len(got), len(base))
			}
			for i, g := range got {
				b := base[i]
				gm, bm := g.Reading.Main, b.Reading.Main
				if !slices.Equal(g.Chord.Intervals(), b.Chord.Intervals()) || g.IsInput != b.IsInput ||
					gm.Kind != bm.Kind || gm.Function != bm.Function || gm.Target != bm.Target || !slices.Equal(gm.Modes, bm.Modes) ||
					!slices.Equal(g.Cautions.Cautions, b.Cautions.Cautions) ||
					!slices.Equal(above(g.Chord.Root(), g.OutOfKey.Notes), above(b.Chord.Root(), b.OutOfKey.Notes)) {
					t.Errorf("%v in %s major: variation %d %v differs from %v in C", moved.Notes(), tonic, i, g.Chord.Notes(), b.Chord.Notes())
				}
			}
		}
	}
}

// above returns the intervals from the root up to the notes.
func above(root material.NoteName, notes []material.NoteName) []material.Interval {
	out := make([]material.Interval, 0, len(notes))
	for _, n := range notes {
		iv, err := material.IntervalBetween(root, n)
		if err != nil {
			panic(err)
		}
		out = append(out, iv)
	}
	return out
}
