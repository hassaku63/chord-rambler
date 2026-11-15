package progression_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

// phraseBeats is four bars of 4/4 (#91, ADR-0006).
const phraseBeats = 16

// crosses reports whether a slot of the lengths, laid one after another from
// beat 0, has a boundary of a phrase strictly inside it, counting beat by
// beat.
func crosses(beats []int) bool {
	start := 0
	for _, b := range beats {
		for beat := start + 1; beat < start+b; beat++ {
			if beat%phraseBeats == 0 {
				return true
			}
		}
		start += b
	}
	return false
}

// checkPhrases checks the nesting of the result: the phrases hold the slots in
// order, every phrase but the last is four bars, and the last is one beat to
// four bars.
func checkPhrases(t *testing.T, p progression.Progression) {
	t.Helper()
	sections := p.Sections()
	if len(sections) != 1 {
		t.Fatalf("Sections() has %d sections; want 1", len(sections))
	}
	phrases := sections[0].Phrases()
	var all []progression.Slot
	for i, ph := range phrases {
		sum := 0
		for _, s := range ph.Slots() {
			sum += s.Beats()
		}
		last := i == len(phrases)-1
		if (!last && sum != phraseBeats) || (last && (sum < 1 || sum > phraseBeats)) {
			t.Errorf("%v: phrase %d of %d is %d beats", lengths(p.Slots()), i, len(phrases), sum)
		}
		all = append(all, ph.Slots()...)
	}
	if !slices.Equal(all, p.Slots()) {
		t.Errorf("the phrases hold %v; want %v", lengths(all), lengths(p.Slots()))
	}
}

// bases returns every progression of up to four slots of the lengths, built by
// appending. A sequence whose append fails with ErrCrossesPhrase is left out;
// any other error fails the test.
func bases(t *testing.T, beats []int) []progression.Progression {
	t.Helper()
	out := []progression.Progression{{}}
	frontier := out
	for range 4 {
		var next []progression.Progression
		for _, p := range frontier {
			for _, b := range beats {
				q, err := p.Append(slot(progression.Unchosen(), b))
				if errors.Is(err, progression.ErrCrossesPhrase) {
					continue
				}
				if err != nil {
					t.Fatalf("Append(%d) after %v: %v", b, lengths(p.Slots()), err)
				}
				next = append(next, q)
			}
		}
		out = append(out, next...)
		frontier = next
	}
	return out
}

type operation struct {
	name string
	do   func(progression.Progression) (progression.Progression, error)
	want func([]progression.Slot) []progression.Slot
}

// operations returns every append, insert, replace and delete on a
// progression of n slots, with a slot of each of the lengths.
func operations(n int, beats []int) []operation {
	var ops []operation
	for _, b := range beats {
		s := slot(progression.NoChord(), b)
		ops = append(ops, operation{"append", func(p progression.Progression) (progression.Progression, error) { return p.Append(s) },
			func(in []progression.Slot) []progression.Slot { return append(in, s) }})
		for i := range n + 1 {
			ops = append(ops, operation{"insert", func(p progression.Progression) (progression.Progression, error) { return p.Insert(i, s) },
				func(in []progression.Slot) []progression.Slot { return slices.Insert(in, i, s) }})
		}
		for i := range n {
			ops = append(ops, operation{"replace", func(p progression.Progression) (progression.Progression, error) { return p.Replace(i, s) },
				func(in []progression.Slot) []progression.Slot { in[i] = s; return in }})
		}
	}
	for i := range n {
		ops = append(ops, operation{"delete", func(p progression.Progression) (progression.Progression, error) { return p.Delete(i) },
			func(in []progression.Slot) []progression.Slot { return slices.Delete(in, i, i+1) }})
	}
	return ops
}

// Every operation within the progression succeeds exactly when no slot of the
// result crosses a phrase, and its result is the slots the operation names,
// nested in phrases. A failure is ErrCrossesPhrase and leaves the progression
// as it was.
func TestOperationsKeepTheBoundaryRule(t *testing.T) {
	beats := []int{1, 3, 4, 6, 8, 16, 17}
	// The test may not pass by every operation succeeding, or by every one
	// failing.
	succeeded, failed := 0, 0
	for _, p := range bases(t, []int{2, 4, 7}) {
		before := p.Slots()
		for _, op := range operations(p.Len(), beats) {
			want := op.want(p.Slots())
			got, err := op.do(p)
			if !slices.Equal(p.Slots(), before) {
				t.Fatalf("%s on %v changed it to %v", op.name, lengths(before), lengths(p.Slots()))
			}
			if crosses(lengths(want)) {
				failed++
				if !errors.Is(err, progression.ErrCrossesPhrase) {
					t.Errorf("%s on %v to %v: err = %v; want ErrCrossesPhrase", op.name, lengths(before), lengths(want), err)
				}
				continue
			}
			succeeded++
			if err != nil || !slices.Equal(got.Slots(), want) {
				t.Errorf("%s on %v = %v, %v; want %v", op.name, lengths(before), lengths(got.Slots()), err, lengths(want))
				continue
			}
			checkPhrases(t, got)
		}
	}
	if succeeded == 0 || failed == 0 {
		t.Errorf("%d operations succeeded and %d failed; want some of each", succeeded, failed)
	}
}

var transposeShapes = []struct{ intervals, bass string }{
	{"", ""}, {"M3 P5", ""}, {"m3 d5 m7", ""}, {"M3 P5 m7 M9", ""}, {"M3 P5", "M3"}, {"M3 P5", "m7"},
}

// transposeCases calls f with a progression in every key of the 35 tonics and
// seven modes, holding a chord of each shape above on one root, no chord and
// none chosen yet. The root steps through the 35 note names from key to key,
// so every root comes up in several keys. Only the failure the materials
// document is skipped: a scale or a chord beyond a double sharp or a double
// flat (material.ErrAlterOutOfRange).
func transposeCases(t *testing.T, f func(p progression.Progression, c material.Chord)) {
	t.Helper()
	var notes []material.NoteName
	for l := material.C; l <= material.B; l++ {
		for a := -2; a <= 2; a++ {
			notes = append(notes, note(l, a))
		}
	}
	keys := 0
	for _, tonic := range notes {
		for mode := material.Ionian; mode <= material.Locrian; mode++ {
			key, err := material.NewScale(tonic, mode)
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatalf("NewScale(%s, %s): %v", tonic, mode, err)
			}
			root := notes[keys%len(notes)]
			keys++
			for _, s := range transposeShapes {
				c, err := material.NewChord(root, ivs(s.intervals)...)
				if err == nil && s.bass != "" {
					c, err = c.WithBass(ivs(s.bass)[0])
				}
				if errors.Is(err, material.ErrAlterOutOfRange) {
					continue
				}
				if err != nil {
					t.Fatalf("chord on %s %v: %v", root, s, err)
				}
				f(transposeProgression(t, key, c), c)
			}
		}
	}
	if keys < len(notes) {
		t.Errorf("%d keys; want at least one for each of the %d roots", keys, len(notes))
	}
}

func transposeProgression(t *testing.T, key material.Scale, c material.Chord) progression.Progression {
	t.Helper()
	unchosen, err := slot(progression.Unchosen(), 4).WithRole(progression.RoleFloat)
	if err != nil {
		t.Fatal(err)
	}
	p := progression.New(key, progression.Meter{})
	for _, s := range []progression.Slot{slot(progression.ChordOf(c), 6), slot(progression.NoChord(), 2), unchosen} {
		if p, err = p.Append(s); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

// allIntervals returns every interval of the five qualities, simple and
// compound.
func allIntervals() []material.Interval {
	var out []material.Interval
	for number := 1; number <= 13; number++ {
		for _, q := range []material.Quality{material.Perfect, material.Major, material.Minor, material.Augmented, material.Diminished} {
			if iv, err := material.NewInterval(number, q); err == nil {
				out = append(out, iv)
			}
		}
	}
	return out
}

func pitchClassPlus(n material.NoteName, semitones int) int {
	return ((n.PitchClass().Int()+semitones)%12 + 12) % 12
}

// Transposing up and back down by the same interval gives the progression
// back. Up, the tonic and the root move by the half steps of the interval,
// the shape, the mode, the lengths and the roles stay, and the slots with no
// chord stay. Only the failure up that the materials document is skipped
// (material.ErrAlterOutOfRange); going back never fails.
func TestTransposeThereAndBack(t *testing.T) {
	ivList := allIntervals()
	back := 0
	transposeCases(t, func(p progression.Progression, c material.Chord) {
		for _, iv := range ivList {
			up, err := p.Transpose(iv)
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatalf("%s in %v up %s: %v", c.Root(), p.Key(), iv, err)
			}
			checkMoved(t, p, up, iv)
			down, err := up.TransposeDown(iv)
			if err != nil || down.Key() != p.Key() || !slices.Equal(down.Slots(), p.Slots()) {
				t.Errorf("%v %v up and down %s = %v %v, %v", p.Key(), p.Slots(), iv, down.Key(), down.Slots(), err)
				continue
			}
			back++
		}
	})
	if back == 0 {
		t.Error("every transposition failed; want some to go there and back")
	}
}

func checkMoved(t *testing.T, from, to progression.Progression, iv material.Interval) {
	t.Helper()
	semitones := iv.Semitones()
	if to.Key().Mode() != from.Key().Mode() || to.Key().Tonic().PitchClass().Int() != pitchClassPlus(from.Key().Tonic(), semitones) {
		t.Errorf("%v up %s is %v", from.Key(), iv, to.Key())
	}
	fs, ts := from.Slots(), to.Slots()
	for i := range fs {
		if fs[i].Beats() != ts[i].Beats() || fs[i].Role() != ts[i].Role() {
			t.Errorf("slot %d up %s: %v; want the length and the role of %v", i, iv, ts[i], fs[i])
		}
		fc, ok := fs[i].Content().Chord()
		if !ok {
			if ts[i].Content() != fs[i].Content() {
				t.Errorf("slot %d up %s: %v; want %v", i, iv, ts[i].Content(), fs[i].Content())
			}
			continue
		}
		tc, ok := ts[i].Content().Chord()
		if !ok || material.ShapeOf(tc) != material.ShapeOf(fc) || tc.Root().PitchClass().Int() != pitchClassPlus(fc.Root(), semitones) {
			t.Errorf("slot %d: %v up %s is %v", i, fc, iv, tc)
		}
	}
}
