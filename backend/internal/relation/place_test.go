package relation_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// built appends slots of the lengths, none chosen yet, to an empty progression
// in 4/4.
func built(t *testing.T, beats ...int) progression.Progression {
	t.Helper()
	p := progression.Progression{}
	for _, b := range beats {
		s, err := progression.NewSlot(progression.Unchosen(), b)
		if err != nil {
			t.Fatalf("NewSlot(%d beats): %v", b, err)
		}
		next, err := p.Append(s)
		if err != nil {
			t.Fatalf("Append(%d beats) after %d slots: %v", b, p.Len(), err)
		}
		p = next
	}
	return p
}

// pl is a Place in the order of its fields, to keep the table short.
func pl(bar, beat, phrase, index int, lastPhrase, lastProg bool) relation.Place {
	return relation.Place{
		Bar: bar, Beat: beat, Phrase: phrase, IndexInPhrase: index,
		LastInPhrase: lastPhrase, LastInProgression: lastProg,
	}
}

// The places of slots, from the cases of #93. The lengths are in beats.
func TestPlaces(t *testing.T) {
	tests := []struct {
		name  string
		beats []int
		want  []relation.Place
	}{
		{"no slots", nil, nil},
		{"one slot", []int{4}, []relation.Place{pl(1, 1, 1, 1, true, true)}},
		{
			"the default: eight bar slots, the fourth and the eighth end a phrase",
			[]int{4, 4, 4, 4, 4, 4, 4, 4},
			[]relation.Place{
				pl(1, 1, 1, 1, false, false), pl(2, 1, 1, 2, false, false),
				pl(3, 1, 1, 3, false, false), pl(4, 1, 1, 4, true, false),
				pl(5, 1, 2, 1, false, false), pl(6, 1, 2, 2, false, false),
				pl(7, 1, 2, 3, false, false), pl(8, 1, 2, 4, true, true),
			},
		},
		{
			"a slot of two bars moves the bar of the next",
			[]int{4, 8, 4},
			[]relation.Place{
				pl(1, 1, 1, 1, false, false), pl(2, 1, 1, 2, false, false),
				pl(4, 1, 1, 3, true, true),
			},
		},
		{
			"two slots of two beats in a bar start on the beats 1 and 3",
			[]int{2, 2, 2, 2},
			[]relation.Place{
				pl(1, 1, 1, 1, false, false), pl(1, 3, 1, 2, false, false),
				pl(2, 1, 1, 3, false, false), pl(2, 3, 1, 4, true, true),
			},
		},
		{
			"a slot that crosses a bar line is placed by where it starts",
			[]int{2, 4, 2},
			[]relation.Place{
				pl(1, 1, 1, 1, false, false), pl(1, 3, 1, 2, false, false),
				pl(2, 3, 1, 3, true, true),
			},
		},
		{
			"a last phrase shorter than four bars ends on its last slot",
			[]int{4, 4, 4, 4, 4, 4},
			[]relation.Place{
				pl(1, 1, 1, 1, false, false), pl(2, 1, 1, 2, false, false),
				pl(3, 1, 1, 3, false, false), pl(4, 1, 1, 4, true, false),
				pl(5, 1, 2, 1, false, false), pl(6, 1, 2, 2, true, true),
			},
		},
		{
			"a slot of four bars fills a phrase, and the next phrase counts again from 1",
			[]int{16, 3, 1, 4},
			[]relation.Place{
				pl(1, 1, 1, 1, true, false),
				pl(5, 1, 2, 1, false, false), pl(5, 4, 2, 2, false, false),
				pl(6, 1, 2, 3, true, true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relation.Places(built(t, tt.beats...))
			if !slices.Equal(got, tt.want) {
				t.Errorf("Places(%v beats) = %v; want %v", tt.beats, got, tt.want)
			}
		})
	}
}

// Places does not change the progression it reads, and its answer does not
// share anything with it.
func TestPlacesLeavesTheProgression(t *testing.T) {
	p := built(t, 4, 4, 4, 4, 4)
	before := p.Slots()
	first := relation.Places(p)
	first[0].Bar = 99
	if again := relation.Places(p); again[0].Bar != 1 {
		t.Errorf("Places after changing an earlier answer: first bar %d; want 1", again[0].Bar)
	}
	if !slices.Equal(p.Slots(), before) || p.Len() != 5 {
		t.Errorf("the progression changed: %d slots", p.Len())
	}
}

// compositions calls f with every list of up to maxSlots lengths in 1..maxBeats
// beats whose progression can be built; lists that put a slot across a phrase
// boundary are skipped, as the progression refuses them.
func compositions(t *testing.T, maxSlots, maxBeats int, f func(beats []int, p progression.Progression)) {
	t.Helper()
	var walk func(beats []int, p progression.Progression)
	walk = func(beats []int, p progression.Progression) {
		f(slices.Clone(beats), p)
		if len(beats) == maxSlots {
			return
		}
		for b := 1; b <= maxBeats; b++ {
			s, err := progression.NewSlot(progression.Unchosen(), b)
			if err != nil {
				t.Fatalf("NewSlot(%d beats): %v", b, err)
			}
			next, err := p.Append(s)
			if errors.Is(err, progression.ErrCrossesPhrase) {
				continue
			}
			if err != nil {
				t.Fatalf("Append(%d beats) after %v: %v", b, beats, err)
			}
			walk(append(beats, b), next)
		}
	}
	walk(nil, progression.Progression{})
}

// For every progression of up to six slots of up to eight beats: there is one
// place per slot; only the last slot ends the progression; the slots of a
// phrase are numbered 1, 2, 3 and so on, and a phrase ends where the next one
// begins; and the start of each slot, counted from its bar and beat, is the sum of the
// beats of the slots before it.
func TestPlacesProperties(t *testing.T) {
	const perBar = 4
	checked := 0
	compositions(t, 6, 8, func(beats []int, p progression.Progression) {
		checked++
		got := relation.Places(p)
		if len(got) != len(beats) {
			t.Fatalf("Places(%v beats): %d places; want %d", beats, len(got), len(beats))
		}
		start := 0
		for i, pl := range got {
			if pl.Bar < 1 || pl.Beat < 1 || pl.Beat > perBar || pl.Phrase < 1 || pl.IndexInPhrase < 1 {
				t.Errorf("Places(%v beats)[%d] = %+v; want every number from 1 and a beat of at most %d",
					beats, i, pl, perBar)
			}
			if at := (pl.Bar-1)*perBar + pl.Beat - 1; at != start {
				t.Errorf("Places(%v beats)[%d] starts at beat %d; the slots before it add up to %d",
					beats, i, at, start)
			}
			start += beats[i]

			if pl.LastInProgression != (i == len(got)-1) {
				t.Errorf("Places(%v beats)[%d].LastInProgression = %v", beats, i, pl.LastInProgression)
			}
			next := i + 1
			switch {
			case next == len(got):
				if !pl.LastInPhrase {
					t.Errorf("Places(%v beats)[%d]: the last slot does not end its phrase", beats, i)
				}
			case got[next].Phrase == pl.Phrase:
				if pl.LastInPhrase || got[next].IndexInPhrase != pl.IndexInPhrase+1 {
					t.Errorf("Places(%v beats)[%d..%d] = %+v, %+v: want the same phrase to go on by one",
						beats, i, next, pl, got[next])
				}
			default:
				if !pl.LastInPhrase || got[next].Phrase != pl.Phrase+1 || got[next].IndexInPhrase != 1 {
					t.Errorf("Places(%v beats)[%d..%d] = %+v, %+v: want a new phrase to start at 1",
						beats, i, next, pl, got[next])
				}
			}
		}
	})
	if checked < 100 {
		t.Errorf("only %d progressions were checked; the test may not pass by checking none", checked)
	}
}
