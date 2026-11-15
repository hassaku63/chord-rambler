package material_test

import (
	"errors"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

func TestNewNoteNameRejects(t *testing.T) {
	// Accidentals reach a double sharp and a double flat (Hutchinson, Music
	// Theory for the 21st-Century Classroom, 1.4 Accidentals:
	// https://musictheory.pugetsound.edu/mt21c/Accidentals.html). Nothing
	// beyond.
	tests := []struct {
		name   string
		letter material.Letter
		alter  int
		want   error
	}{
		{name: "triple sharp", letter: material.F, alter: 3, want: material.ErrAlterOutOfRange},
		{name: "triple flat", letter: material.B, alter: -3, want: material.ErrAlterOutOfRange},
		{name: "letter past B", letter: material.B + 1, alter: 0, want: material.ErrInvalidLetter},
		{name: "letter before C", letter: material.C - 1, alter: 0, want: material.ErrInvalidLetter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := material.NewNoteName(tt.letter, tt.alter)
			if !errors.Is(err, tt.want) {
				t.Errorf("NewNoteName(%d, %d) = %v, %v; want error %v", int(tt.letter), tt.alter, got, err, tt.want)
			}
		})
	}
}

func TestNoteNamePitchClass(t *testing.T) {
	// A sharp adds a half step and a flat takes one; doubles add or take two
	// (Hutchinson 1.4 Accidentals, source above). The natural rows are derived,
	// not stated on a page: C up to each natural is the perfect or major
	// interval of that number, with the half steps of TestIntervalSemitones.
	// Enharmonic names share a pitch class.
	tests := []struct {
		note material.NoteName
		want int
	}{
		{note(material.C, nat), 0},
		{note(material.D, nat), 2},
		{note(material.E, nat), 4},
		{note(material.F, nat), 5},
		{note(material.G, nat), 7},
		{note(material.A, nat), 9},
		{note(material.B, nat), 11},
		{note(material.F, sharp), 6},
		{note(material.G, flat), 6},
		{note(material.B, sharp), 0},
		{note(material.C, flat), 11},
		{note(material.D, flat2), 0},
		{note(material.B, sharp2), 1},
		{note(material.C, flat2), 10},
	}
	for _, tt := range tests {
		t.Run(tt.note.String(), func(t *testing.T) {
			if got := tt.note.PitchClass().Int(); got != tt.want {
				t.Errorf("%s.PitchClass() = %d, want %d", tt.note, got, tt.want)
			}
		})
	}
}
