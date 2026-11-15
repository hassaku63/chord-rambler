package progression_test

import (
	"errors"
	"slices"
	"testing"

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
