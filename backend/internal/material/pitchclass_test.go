package material_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

func pcs(ns ...int) material.PitchClassSet {
	members := make([]material.PitchClass, 0, len(ns))
	for _, n := range ns {
		pc, err := material.NewPitchClass(n)
		if err != nil {
			panic(err)
		}
		members = append(members, pc)
	}
	return material.NewPitchClassSet(members...)
}

func ints(s material.PitchClassSet) []int {
	var out []int
	for _, pc := range s.PitchClasses() {
		out = append(out, pc.Int())
	}
	return out
}

func TestNewPitchClassRejects(t *testing.T) {
	for _, n := range []int{-1, 12} {
		if got, err := material.NewPitchClass(n); !errors.Is(err, material.ErrPitchClassOutOfRange) {
			t.Errorf("NewPitchClass(%d) = %v, %v; want error %v", n, got, err, material.ErrPitchClassOutOfRange)
		}
	}
}

func TestPitchClassSetTranspose(t *testing.T) {
	tests := []struct {
		name      string
		set       material.PitchClassSet
		semitones int
		want      []int
	}{
		{name: "up a whole step", set: pcs(0, 4, 7), semitones: 2, want: []int{2, 6, 9}},
		{name: "wraps past 11", set: pcs(0, 4, 7), semitones: 7, want: []int{2, 7, 11}},
		{name: "down", set: pcs(0, 4, 7), semitones: -1, want: []int{3, 6, 11}},
		{name: "an octave", set: pcs(0, 4, 7), semitones: 12, want: []int{0, 4, 7}},
		{name: "empty", set: material.PitchClassSet{}, semitones: 5, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ints(tt.set.Transpose(tt.semitones)); !slices.Equal(got, tt.want) {
				t.Errorf("%s.Transpose(%d) = %v, want %v", tt.set, tt.semitones, got, tt.want)
			}
		})
	}
}

func TestPitchClassSetMembership(t *testing.T) {
	cMajor := pcs(0, 2, 4, 5, 7, 9, 11)
	triad := pcs(0, 4, 7)
	if !triad.IsSubsetOf(cMajor) {
		t.Errorf("%s.IsSubsetOf(%s) = false, want true", triad, cMajor)
	}
	if cMajor.IsSubsetOf(triad) {
		t.Errorf("%s.IsSubsetOf(%s) = true, want false", cMajor, triad)
	}
	if !(material.PitchClassSet{}).IsSubsetOf(triad) {
		t.Errorf("the empty set is not a subset of %s", triad)
	}
	if !cMajor.Contains(note(material.F, nat).PitchClass()) {
		t.Errorf("%s does not contain F", cMajor)
	}
	if cMajor.Contains(note(material.F, sharp).PitchClass()) {
		t.Errorf("%s contains F♯", cMajor)
	}
	// No spelling: E♭ and D♯ are the same member.
	if got := ints(pcs(0, 3).Union(material.NewPitchClassSet(note(material.D, sharp).PitchClass()))); !slices.Equal(got, []int{0, 3}) {
		t.Errorf("{0 3} ∪ {D♯} = %v, want [0 3]", got)
	}
	if got := ints(pcs(0, 4).Union(pcs(4, 7))); !slices.Equal(got, []int{0, 4, 7}) {
		t.Errorf("{0 4} ∪ {4 7} = %v, want [0 4 7]", got)
	}
}
