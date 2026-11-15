package material_test

import (
	"errors"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

const (
	P = material.Perfect
	M = material.Major
	m = material.Minor
	A = material.Augmented
	d = material.Diminished
)

func TestIntervalSemitones(t *testing.T) {
	// Sources (Hutchinson, Music Theory for the 21st-Century Classroom):
	//   - 5.1: unisons, fourths, fifths and octaves are perfect; seconds,
	//     thirds, sixths and sevenths are major or minor.
	//     https://musictheory.pugetsound.edu/mt21c/IntervalsIntroduction.html
	//   - 5.2: half steps of m2 (1) through P8 (12).
	//     https://musictheory.pugetsound.edu/mt21c/HowToIdentifyIntervals.html
	//   - 5.5: augmented is a half step larger than perfect or major,
	//     diminished a half step smaller than perfect or minor.
	//     https://musictheory.pugetsound.edu/mt21c/AugmentedAndDiminishedIntervals.html
	//   - 31.1: 9 is 2, 11 is 4 and 13 is 6 an octave higher.
	//     https://musictheory.pugetsound.edu/mt21c/JazzChordBasics.html
	// Derived from those definitions, not stated on the pages: P1 = 0, and the
	// 10th and 12th as the 3rd and 5th an octave higher.
	// A convention of this project, not a cited fact: the diminished unison
	// (-1) exists, as the inverse of the augmented unison, so that moving down
	// by a half step on the same letter (C♯ to C) has a name. Some textbooks
	// say a unison cannot be diminished (docs/architecture/domain-model.md).
	tests := []struct {
		number  int
		quality material.Quality
		want    int
	}{
		{1, d, -1}, {1, P, 0}, {1, A, 1},
		{2, d, 0}, {2, m, 1}, {2, M, 2}, {2, A, 3},
		{3, d, 2}, {3, m, 3}, {3, M, 4}, {3, A, 5},
		{4, d, 4}, {4, P, 5}, {4, A, 6},
		{5, d, 6}, {5, P, 7}, {5, A, 8},
		{6, d, 7}, {6, m, 8}, {6, M, 9}, {6, A, 10},
		{7, d, 9}, {7, m, 10}, {7, M, 11}, {7, A, 12},
		{8, d, 11}, {8, P, 12}, {8, A, 13},
		{9, d, 12}, {9, m, 13}, {9, M, 14}, {9, A, 15},
		{10, d, 14}, {10, m, 15}, {10, M, 16}, {10, A, 17},
		{11, d, 16}, {11, P, 17}, {11, A, 18},
		{12, d, 18}, {12, P, 19}, {12, A, 20},
		{13, d, 19}, {13, m, 20}, {13, M, 21}, {13, A, 22},
	}
	for _, tt := range tests {
		iv := interval(tt.number, tt.quality)
		t.Run(iv.String(), func(t *testing.T) {
			if got := iv.Semitones(); got != tt.want {
				t.Errorf("%s.Semitones() = %d, want %d", iv, got, tt.want)
			}
		})
	}
	if got, want := len(allIntervals()), len(tests); got != want {
		t.Errorf("NewInterval accepts %d intervals, the table has %d", got, want)
	}
}

func TestNewIntervalRejects(t *testing.T) {
	// 5.1 and 5.5 (sources above): perfect numbers are never major or minor,
	// the others never perfect.
	tests := []struct {
		name    string
		number  int
		quality material.Quality
	}{
		{name: "major fifth", number: 5, quality: M},
		{name: "minor unison", number: 1, quality: m},
		{name: "minor eleventh", number: 11, quality: m},
		{name: "perfect third", number: 3, quality: P},
		{name: "perfect ninth", number: 9, quality: P},
		{name: "number zero", number: 0, quality: P},
		{name: "fourteenth", number: 14, quality: M},
		{name: "unknown quality", number: 3, quality: material.Quality(9)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := material.NewInterval(tt.number, tt.quality)
			if !errors.Is(err, material.ErrInvalidInterval) {
				t.Errorf("NewInterval(%d, %s) = %v, %v; want error %v", tt.number, tt.quality, got, err, material.ErrInvalidInterval)
			}
		})
	}
}

func TestIntervalBetween(t *testing.T) {
	// The number comes from the letters, the quality from the half steps
	// (Hutchinson 5.2 and 5.5, sources above). The rows are derived from those
	// definitions.
	tests := []struct {
		lower, upper material.NoteName
		want         material.Interval
	}{
		{note(material.C, nat), note(material.E, flat), interval(3, m)},
		{note(material.C, nat), note(material.D, sharp), interval(2, A)},
		{note(material.C, nat), note(material.E, nat), interval(3, M)},
		{note(material.C, nat), note(material.G, nat), interval(5, P)},
		{note(material.C, nat), note(material.B, nat), interval(7, M)},
		{note(material.E, nat), note(material.F, nat), interval(2, m)},
		{note(material.F, nat), note(material.B, nat), interval(4, A)},
		{note(material.B, nat), note(material.F, nat), interval(5, d)},
		{note(material.B, nat), note(material.C, nat), interval(2, m)},
		{note(material.A, nat), note(material.C, nat), interval(3, m)},
		{note(material.D, nat), note(material.C, nat), interval(7, m)},
		{note(material.B, flat), note(material.F, nat), interval(5, P)},
		{note(material.C, sharp), note(material.B, flat), interval(7, d)},
		{note(material.C, nat), note(material.C, nat), interval(1, P)},
		{note(material.C, nat), note(material.C, sharp), interval(1, A)},
		{note(material.C, nat), note(material.C, flat), interval(1, d)},
	}
	for _, tt := range tests {
		t.Run(tt.lower.String()+"-"+tt.upper.String(), func(t *testing.T) {
			got, err := material.IntervalBetween(tt.lower, tt.upper)
			if err != nil || got != tt.want {
				t.Errorf("IntervalBetween(%s, %s) = %v, %v; want %s", tt.lower, tt.upper, got, err, tt.want)
			}
		})
	}
}

func TestIntervalBetweenRejects(t *testing.T) {
	// Distances beyond augmented and diminished have none of the five
	// qualities.
	tests := []struct {
		lower, upper material.NoteName
	}{
		{note(material.C, sharp), note(material.E, flat2)}, // a doubly diminished third
		{note(material.C, flat), note(material.G, sharp)},  // a doubly augmented fifth
		{note(material.B, sharp), note(material.C, flat)},  // a second of -1 half step
	}
	for _, tt := range tests {
		t.Run(tt.lower.String()+"-"+tt.upper.String(), func(t *testing.T) {
			got, err := material.IntervalBetween(tt.lower, tt.upper)
			if !errors.Is(err, material.ErrInvalidInterval) {
				t.Errorf("IntervalBetween(%s, %s) = %v, %v; want error %v", tt.lower, tt.upper, got, err, material.ErrInvalidInterval)
			}
		})
	}
}
