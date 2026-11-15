package progression_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

func slotsOf(c progression.Content, beats []int) []progression.Slot {
	out := make([]progression.Slot, 0, len(beats))
	for _, b := range beats {
		out = append(out, slot(c, b))
	}
	return out
}

// A range of slots is replaced in one step by slots of the same total length,
// so a slot of any phrase can be split, joined or given other lengths (#96).
// A failure leaves the progression as it was.
func TestReplaceRange(t *testing.T) {
	eight := []int{4, 4, 4, 4, 4, 4, 4, 4}
	tests := []struct {
		name    string
		beats   []int
		i, j    int
		with    []int
		want    []int
		wantErr error
	}{
		{"split slot 0 into 2 + 2", eight, 0, 1, []int{2, 2}, []int{2, 2, 4, 4, 4, 4, 4, 4, 4}, nil},
		{"join slots 0 and 1 into 8", eight, 0, 2, []int{8}, []int{8, 4, 4, 4, 4, 4, 4}, nil},
		{"share slots 0 and 1 out as 6 and 2", eight, 0, 2, []int{6, 2}, []int{6, 2, 4, 4, 4, 4, 4, 4}, nil},
		{"split bar 4 into 3 + 1, up to the boundary", eight, 3, 4, []int{3, 1}, []int{4, 4, 4, 3, 1, 4, 4, 4, 4}, nil},
		{"split the last slot", eight, 7, 8, []int{1, 3}, []int{4, 4, 4, 4, 4, 4, 4, 1, 3}, nil},
		{"a range across the boundary, given back on both sides", eight, 3, 5, []int{1, 3, 3, 1}, []int{4, 4, 4, 1, 3, 3, 1, 4, 4, 4}, nil},
		{"a range in the middle of slots of different lengths", []int{1, 3, 4, 4, 4}, 1, 3, []int{5, 2}, []int{1, 5, 2, 4, 4}, nil},
		// Bars 4 and 5 as 2 + 4 + 2: the 4 starts on beat 3 of bar 4.
		{"a slot given across the boundary", eight, 3, 5, []int{2, 4, 2}, nil, progression.ErrCrossesPhrase},
		{"join bars 4 and 5", eight, 3, 5, []int{8}, nil, progression.ErrCrossesPhrase},
		{"fewer beats", eight, 0, 2, []int{4, 3}, nil, progression.ErrRangeLength},
		{"more beats", eight, 0, 1, []int{4, 1}, nil, progression.ErrRangeLength},
		{"no slots given", eight, 0, 1, nil, nil, progression.ErrRangeLength},
		{"an empty range", eight, 2, 2, nil, nil, progression.ErrIndex},
		{"j before i", eight, 3, 2, []int{4}, nil, progression.ErrIndex},
		{"i before the start", eight, -1, 1, []int{4}, nil, progression.ErrIndex},
		{"j after the end", eight, 7, 9, []int{4}, nil, progression.ErrIndex},
		{"no slots at all", nil, 0, 1, []int{4}, nil, progression.ErrIndex},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := build(t, tt.beats...)
			got, err := p.ReplaceRange(tt.i, tt.j, slotsOf(progression.NoChord(), tt.with)...)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("err = %v; want %v", err, tt.wantErr)
				}
			} else if err != nil || !slices.Equal(lengths(got.Slots()), tt.want) {
				t.Errorf("slots = %v, %v; want %v", lengths(got.Slots()), err, tt.want)
			}
			if !slices.Equal(lengths(p.Slots()), tt.beats) {
				t.Errorf("the original is now %v; want %v", lengths(p.Slots()), tt.beats)
			}
		})
	}
}

// starts returns the beat each slot starts on, adding up the lengths before
// it.
func starts(slots []progression.Slot) []int {
	out := make([]int, 0, len(slots))
	beat := 0
	for _, s := range slots {
		out = append(out, beat)
		beat += s.Beats()
	}
	return out
}

// replacements returns ways to fill the beats: one slot, every split into two,
// all single beats, and one beat too few and too many.
func replacements(beats int) [][]int {
	out := [][]int{{beats}, {beats + 1}, slices.Repeat([]int{1}, beats)}
	if beats > 1 {
		out = append(out, []int{beats - 1})
	}
	for a := 1; a < beats; a++ {
		out = append(out, []int{a, beats - a})
	}
	return out
}

// Replacing any range of any base fails with ErrRangeLength exactly when the
// lengths do not add up, and otherwise with ErrCrossesPhrase exactly when a
// slot crosses a phrase. A success holds the slots before the range, the slots
// given and the slots after it; the slots outside the range start on the
// same beats as before; and the phrases nest the slots. A failure leaves the
// progression as it was.
func TestReplaceRangeKeepsTheOtherSlots(t *testing.T) {
	// The test may not pass by every case taking one path.
	succeeded, crossed, mismatched := 0, 0, 0
	for _, p := range bases(t, []int{2, 4, 7}) {
		before := p.Slots()
		for i := range before {
			for j := i + 1; j <= len(before); j++ {
				have := 0
				for _, s := range before[i:j] {
					have += s.Beats()
				}
				for _, with := range replacements(have) {
					given := slotsOf(progression.NoChord(), with)
					want := slices.Concat(before[:i], given, before[j:])
					got, err := p.ReplaceRange(i, j, given...)
					if !slices.Equal(p.Slots(), before) {
						t.Fatalf("ReplaceRange(%d, %d) on %v changed it to %v", i, j, lengths(before), lengths(p.Slots()))
					}
					give := 0
					for _, b := range with {
						give += b
					}
					switch {
					case give != have:
						mismatched++
						if !errors.Is(err, progression.ErrRangeLength) {
							t.Errorf("ReplaceRange(%d, %d, %v) on %v: err = %v; want ErrRangeLength", i, j, with, lengths(before), err)
						}
					case crosses(lengths(want)):
						crossed++
						if !errors.Is(err, progression.ErrCrossesPhrase) {
							t.Errorf("ReplaceRange(%d, %d, %v) on %v: err = %v; want ErrCrossesPhrase", i, j, with, lengths(before), err)
						}
					default:
						succeeded++
						checkReplaced(t, before, got, err, i, j, given, want)
					}
				}
			}
		}
	}
	if succeeded == 0 || crossed == 0 || mismatched == 0 {
		t.Errorf("%d succeeded, %d crossed and %d did not add up; want some of each", succeeded, crossed, mismatched)
	}
}

func checkReplaced(t *testing.T, before []progression.Slot, got progression.Progression, err error, i, j int, given, want []progression.Slot) {
	t.Helper()
	if err != nil || !slices.Equal(got.Slots(), want) {
		t.Errorf("ReplaceRange(%d, %d, %v) on %v = %v, %v; want %v", i, j, lengths(given), lengths(before), lengths(got.Slots()), err, lengths(want))
		return
	}
	was, now := starts(before), starts(got.Slots())
	after := len(now) - (len(before) - j) // the index of the first slot after the range, now
	if !slices.Equal(now[:i], was[:i]) || !slices.Equal(now[after:], was[j:]) {
		t.Errorf("ReplaceRange(%d, %d, %v) on %v: the other slots start on %v, %v; want %v, %v",
			i, j, lengths(given), lengths(before), now[:i], now[after:], was[:i], was[j:])
	}
	checkPhrases(t, got)
}
