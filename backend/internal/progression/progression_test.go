package progression_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

// build appends slots of the lengths, none chosen yet, to an empty progression
// in C major.
func build(t *testing.T, beats ...int) progression.Progression {
	t.Helper()
	p := progression.Progression{}
	for _, b := range beats {
		q, err := p.Append(slot(progression.Unchosen(), b))
		if err != nil {
			t.Fatalf("Append(%d beats) after %v: %v", b, lengths(p.Slots()), err)
		}
		p = q
	}
	return p
}

func lengths(slots []progression.Slot) []int {
	out := make([]int, 0, len(slots))
	for _, s := range slots {
		out = append(out, s.Beats())
	}
	return out
}

// phraseLengths returns the lengths of the slots of each phrase of the one
// section.
func phraseLengths(t *testing.T, p progression.Progression) [][]int {
	t.Helper()
	sections := p.Sections()
	if len(sections) != 1 {
		t.Fatalf("Sections() has %d sections; want 1", len(sections))
	}
	var out [][]int
	for _, ph := range sections[0].Phrases() {
		out = append(out, lengths(ph.Slots()))
	}
	return out
}

// Phrases are cut every four bars from the start, and the last one may be
// shorter (ADR-0006). A slot may cross a bar (#91).
func TestPhrases(t *testing.T) {
	tests := []struct {
		name  string
		beats []int
		want  [][]int
	}{
		{"no slots", nil, nil},
		{"the default: eight slots of one bar", []int{4, 4, 4, 4, 4, 4, 4, 4}, [][]int{{4, 4, 4, 4}, {4, 4, 4, 4}}},
		{"a slot of two bars", []int{8, 4, 4, 4}, [][]int{{8, 4, 4}, {4}}},
		{"a slot of four bars fills a phrase", []int{16, 4}, [][]int{{16}, {4}}},
		{"a last phrase of two bars", []int{4, 4, 4, 4, 4, 4}, [][]int{{4, 4, 4, 4}, {4, 4}}},
		{"a last phrase of one beat", []int{4, 4, 4, 4, 1}, [][]int{{4, 4, 4, 4}, {1}}},
		{"slots across the bars inside a phrase", []int{3, 3, 3, 3, 4, 2}, [][]int{{3, 3, 3, 3, 4}, {2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phraseLengths(t, build(t, tt.beats...)); !slices.EqualFunc(got, tt.want, slices.Equal) {
				t.Errorf("phrases = %v; want %v", got, tt.want)
			}
		})
	}
}

// A slot may not cross the boundary of a phrase (#91). The operation fails
// and the progression it was called on stays as it was.
func TestOperationsKeepTheBoundaries(t *testing.T) {
	four := slot(progression.Unchosen(), 4)
	tests := []struct {
		name  string
		beats []int
		op    func(progression.Progression) (progression.Progression, error)
		want  []int // nil for ErrCrossesPhrase
	}{
		// Bar 4, beat 3 is beat 15 of the phrase; four beats from there end
		// on beat 2 of bar 5.
		{"append four beats from bar 4, beat 3", []int{4, 4, 4, 2}, func(p progression.Progression) (progression.Progression, error) {
			return p.Append(four)
		}, nil},
		{"append two beats from bar 4, beat 3", []int{4, 4, 4, 2}, func(p progression.Progression) (progression.Progression, error) {
			return p.Append(slot(progression.Unchosen(), 2))
		}, []int{4, 4, 4, 2, 2}},
		{"insert at the start, moving a later slot across", []int{4, 4, 4, 4, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Insert(0, slot(progression.Unchosen(), 2))
		}, nil},
		{"insert a bar at the start", []int{4, 4, 4, 4, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Insert(0, four)
		}, []int{4, 4, 4, 4, 4, 4}},
		{"insert at the end", []int{4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Insert(1, slot(progression.Unchosen(), 2))
		}, []int{4, 2}},
		{"replace with a longer slot", []int{4, 4, 4, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Replace(3, slot(progression.Unchosen(), 6))
		}, nil},
		{"replace the first bar with two beats, moving a later slot across", []int{4, 4, 4, 4, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Replace(0, slot(progression.Unchosen(), 2))
		}, nil},
		{"replace with a slot of two bars", []int{4, 4, 4, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Replace(2, slot(progression.Unchosen(), 8))
		}, []int{4, 4, 8, 4}},
		{"delete, moving a later slot across", []int{2, 4, 4, 4, 2, 8}, func(p progression.Progression) (progression.Progression, error) {
			return p.Delete(4)
		}, nil},
		{"delete a slot in the middle", []int{1, 2, 3, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Delete(1)
		}, []int{1, 3, 4}},
		{"delete the first bar", []int{4, 4, 4, 4, 4}, func(p progression.Progression) (progression.Progression, error) {
			return p.Delete(0)
		}, []int{4, 4, 4, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := build(t, tt.beats...)
			got, err := tt.op(p)
			if tt.want == nil {
				if !errors.Is(err, progression.ErrCrossesPhrase) {
					t.Errorf("err = %v (slots %v); want ErrCrossesPhrase", err, lengths(got.Slots()))
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

func TestOperationsOutsideTheProgression(t *testing.T) {
	p := build(t, 4, 4)
	s := slot(progression.NoChord(), 4)
	tests := []struct {
		name string
		op   func() (progression.Progression, error)
	}{
		{"replace -1", func() (progression.Progression, error) { return p.Replace(-1, s) }},
		{"replace 2", func() (progression.Progression, error) { return p.Replace(2, s) }},
		{"insert -1", func() (progression.Progression, error) { return p.Insert(-1, s) }},
		{"insert 3", func() (progression.Progression, error) { return p.Insert(3, s) }},
		{"delete -1", func() (progression.Progression, error) { return p.Delete(-1) }},
		{"delete 2", func() (progression.Progression, error) { return p.Delete(2) }},
		{"delete from no slots", func() (progression.Progression, error) { return progression.Progression{}.Delete(0) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.op(); !errors.Is(err, progression.ErrIndex) {
				t.Errorf("err = %v; want ErrIndex", err)
			}
		})
	}
}

// The operations put the slot where they say, and two operations on the same
// progression do not see each other.
func TestOperationsReturnNewProgressions(t *testing.T) {
	a, b, c := slot(progression.NoChord(), 1), slot(progression.Unchosen(), 2), slot(progression.ChordOf(material.Chord{}), 3)
	p := build(t, 4, 4, 4)
	p, err := p.Delete(2) // a slice with room after its end
	if err != nil {
		t.Fatal(err)
	}
	first, err1 := p.Append(a)
	second, err2 := p.Append(b)
	inserted, err3 := p.Insert(1, c)
	replaced, err4 := p.Replace(0, c)
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		t.Fatal(err)
	}
	x := slot(progression.Unchosen(), 4)
	checks := []struct {
		name string
		got  []progression.Slot
		want []progression.Slot
	}{
		{"the original", p.Slots(), []progression.Slot{x, x}},
		{"append a", first.Slots(), []progression.Slot{x, x, a}},
		{"append b", second.Slots(), []progression.Slot{x, x, b}},
		{"insert c before 1", inserted.Slots(), []progression.Slot{x, c, x}},
		{"replace 0 with c", replaced.Slots(), []progression.Slot{c, x}},
	}
	for _, ch := range checks {
		if !slices.Equal(ch.got, ch.want) {
			t.Errorf("%s: slots = %v; want %v", ch.name, ch.got, ch.want)
		}
	}
	// Changing what Slots returned does not change the progression.
	got := p.Slots()
	got[0] = a
	if p.Slots()[0] != x {
		t.Error("changing the result of Slots changed the progression")
	}
}

// The zero value is C major, 4/4, with no slots; New keeps the key and the
// meter.
func TestNew(t *testing.T) {
	var zero progression.Progression
	if zero.Key() != (material.Scale{}) || zero.Meter() != (progression.Meter{}) || zero.Len() != 0 {
		t.Errorf("zero Progression = %v, %v, %d slots; want C Ionian, 4/4, none", zero.Key(), zero.Meter(), zero.Len())
	}
	key, err := material.NewScale(note(material.E, flat), material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	if p := progression.New(key, progression.Meter{}); p.Key() != key || p.Len() != 0 {
		t.Errorf("New(E♭ Aeolian) = %v, %d slots; want E♭ Aeolian, none", p.Key(), p.Len())
	}
}
