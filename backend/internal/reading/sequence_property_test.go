package reading_test

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// tertianChord is the owner's call on #103 said again: a triad with a third,
// with or without a seventh, and no second, fourth or sixth member.
func tertianChord(ch material.Chord) bool {
	switch ch.Triad() {
	case material.TriadMajor, material.TriadMinor, material.TriadDiminished, material.TriadAugmented:
	default:
		return false
	}
	for _, iv := range ch.Intervals() {
		if n := iv.Number(); n == 2 || n == 4 || n == 6 {
			return false
		}
	}
	return true
}

// degreeIn returns the degree of a chord read as diatonic, or 0.
func degreeIn(t *testing.T, ch material.Chord, key material.Scale) int {
	t.Helper()
	r, err := reading.ReadChord(ch, key)
	if err != nil {
		t.Fatal(err)
	}
	if r.Main.Kind != reading.Diatonic {
		return 0
	}
	pos, err := relation.PositionOf(ch, key)
	if err != nil {
		t.Fatal(err)
	}
	return pos.Degree()
}

// For the chords of the shapes (and the sixth chord) on a root of at most a
// sharp or a flat, in C major, each dominant followed by every chord and each
// other chord followed by a few: a resolution is said of V or a secondary
// dominant only, and it is expected only when the next chord is read as
// diatonic, is tertian and is on the degree pointed to (1 for V). This checks
// that the fact said matches the next chord (docs/architecture/explanations.md,
// error 2).
func TestExpectedOnlyOnTheDegreePointedTo(t *testing.T) {
	key := major(cNat)
	var chords []material.Chord
	for l := material.C; l <= material.B; l++ {
		for a := flat; a <= sharp; a++ {
			for _, s := range slices.Concat(shapes, []struct{ intervals, bass string }{{"M3 P5 M6", ""}}) {
				ch, err := material.NewChord(note(l, a), ivs(s.intervals)...)
				if err == nil && s.bass != "" {
					ch, err = ch.WithBass(ivs(s.bass)[0])
				}
				if errors.Is(err, material.ErrAlterOutOfRange) {
					continue // beyond a double flat, as eachChord skips
				}
				if err != nil {
					t.Fatal(err)
				}
				chords = append(chords, ch)
			}
		}
	}
	target := make([]int, len(chords)) // the degree pointed to, or 0
	degree := make([]int, len(chords))
	for i, ch := range chords {
		r, err := reading.ReadChord(ch, key)
		if err != nil {
			t.Fatal(err)
		}
		degree[i] = degreeIn(t, ch, key)
		s := ch.Seventh()
		switch {
		case r.Main.Kind == reading.SecondaryDominant:
			target[i] = r.Main.Target
		case degree[i] == 5 && tertianChord(ch) && ch.Triad() == material.TriadMajor &&
			(s == material.NoSeventh || s == material.SeventhMinor):
			target[i] = 1
		}
	}
	seen := map[reading.Resolution]int{}
	// After a chord that is not a dominant, a few chords stand for the rest.
	few := []int{index(t, chords, chord(cNat, maj, "")), index(t, chords, g7), index(t, chords, am)}
	for i, prev := range chords {
		nexts := few
		if target[i] != 0 {
			nexts = nil
			for j := range chords {
				nexts = append(nexts, j)
			}
		}
		for _, j := range nexts {
			next := chords[j]
			got := readProgression(t, build(key, bar(prev), bar(next)))[0].Resolution
			seen[got]++
			if got != reading.NoResolution && target[i] == 0 {
				t.Errorf("%v → %v: %s, but the first is not a dominant", prev, next, got)
			}
			if got == reading.ResolutionExpected && (!tertianChord(next) || degree[j] != target[i]) {
				t.Errorf("%v → %v: expected, but the next is not a tertian chord on degree %d", prev, next, target[i])
			}
		}
	}
	for _, r := range []reading.Resolution{reading.NoResolution, reading.ResolutionExpected, reading.ResolutionIrregular} {
		if seen[r] == 0 {
			t.Errorf("no pair read as %s; the checks are vacuous", r)
		}
	}
}

// For every progression of up to four slots of up to eight beats, every slot
// holding V (G in C): a cadence is named on the last slot of a full phrase,
// and on no other slot.
func TestCadenceOnlyAtTheEndOfFullPhrases(t *testing.T) {
	key := major(cNat)
	var walk func(p progression.Progression, depth int)
	walk = func(p progression.Progression, depth int) {
		got := readProgression(t, p)
		places := relation.Places(p)
		end := 0
		for i, s := range p.Slots() {
			end += s.Beats()
			want := reading.NoCadence
			if places[i].LastInPhrase && end%16 == 0 {
				want = reading.HalfCadence
			}
			if got[i].Cadence != want {
				t.Errorf("%d slots, slot %d (bar %d, ends on beat %d): %s; want %s",
					p.Len(), i, places[i].Bar, end, got[i].Cadence, want)
			}
		}
		if depth == 4 {
			return
		}
		for b := 1; b <= 8; b++ {
			s, err := progression.NewSlot(progression.ChordOf(chord(gNat, maj, "")), b)
			if err != nil {
				t.Fatal(err)
			}
			next, err := p.Append(s)
			if errors.Is(err, progression.ErrCrossesPhrase) {
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			walk(next, depth+1)
		}
	}
	walk(progression.New(key, progression.Meter{}), 0)
}

// For every progression of four bars whose slots hold no chord, none chosen
// yet, or one of a few chords of C major, a slot without a chord reads as the
// zero value: nothing is said of it.
func TestSlotsWithoutChordsSayNothing(t *testing.T) {
	contents := []slot{noChord, unchosen, bar(g7), bar(c), bar(d7), bar(f)}
	var walk func(slots []slot)
	walk = func(slots []slot) {
		if len(slots) == 4 {
			for i, r := range readProgression(t, build(major(cNat), slots...)) {
				if _, ok := slots[i].content.Chord(); !ok && !reflect.DeepEqual(r, reading.SlotReading{}) {
					t.Errorf("slot %d of %v: %+v; want the zero value", i, slots, r)
				}
			}
			return
		}
		for _, s := range contents {
			walk(append(slots[:len(slots):len(slots)], s))
		}
	}
	walk(nil)
}

func index(t *testing.T, chords []material.Chord, ch material.Chord) int {
	t.Helper()
	if i := slices.Index(chords, ch); i >= 0 {
		return i
	}
	t.Fatalf("%v is not among the chords", ch)
	return 0
}
