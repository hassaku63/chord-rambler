package material_test

import (
	"errors"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

func TestTranspose(t *testing.T) {
	// The letter moves by the interval number and the alteration makes up the
	// half steps (Hutchinson 5.3, writing a m6 above A as F from A major's F♯:
	// https://musictheory.pugetsound.edu/mt21c/HowToWriteIntervals.html).
	// The rows are derived from that method and the half steps in
	// TestIntervalSemitones.
	tests := []struct {
		from material.NoteName
		by   material.Interval
		want material.NoteName
	}{
		{note(material.E, nat), interval(3, M), note(material.G, sharp)},
		{note(material.B, flat), interval(5, P), note(material.F, nat)},
		{note(material.F, sharp), interval(3, M), note(material.A, sharp)},
		{note(material.B, nat), interval(3, M), note(material.D, sharp)},
		{note(material.A, nat), interval(6, m), note(material.F, nat)},
		{note(material.D, sharp), interval(3, M), note(material.F, sharp2)},
		{note(material.C, flat), interval(3, m), note(material.E, flat2)},
		{note(material.C, nat), interval(9, M), note(material.D, nat)},
		{note(material.C, nat), interval(11, A), note(material.F, sharp)},
		{note(material.C, nat), interval(13, M), note(material.A, nat)},
	}
	for _, tt := range tests {
		t.Run(tt.from.String()+"+"+tt.by.String(), func(t *testing.T) {
			got, err := tt.from.Transpose(tt.by)
			if err != nil || got != tt.want {
				t.Errorf("%s.Transpose(%s) = %v, %v; want %s", tt.from, tt.by, got, err, tt.want)
			}
		})
	}
}

func TestTransposeRejects(t *testing.T) {
	tests := []struct {
		from material.NoteName
		by   material.Interval
		up   bool
	}{
		{note(material.D, sharp2), interval(3, M), true},  // F triple sharp
		{note(material.C, flat2), interval(3, m), true},   // E triple flat
		{note(material.E, flat2), interval(2, A), false},  // D triple flat
		{note(material.E, sharp2), interval(4, d), false}, // B triple sharp
	}
	for _, tt := range tests {
		t.Run(tt.from.String()+" "+tt.by.String(), func(t *testing.T) {
			transpose := tt.from.TransposeDown
			if tt.up {
				transpose = tt.from.Transpose
			}
			got, err := transpose(tt.by)
			if !errors.Is(err, material.ErrAlterOutOfRange) {
				t.Errorf("%s moved by %s (up %t) = %v, %v; want error %v", tt.from, tt.by, tt.up, got, err, material.ErrAlterOutOfRange)
			}
		})
	}
}

// Every note name and every interval: no generator is needed, and the inputs
// are exactly the range the package accepts.

func TestTransposeKeepsPitchClass(t *testing.T) {
	checked := 0
	for _, n := range allNotes() {
		for _, iv := range allIntervals() {
			got, err := n.Transpose(iv)
			if err != nil {
				continue
			}
			checked++
			want := ((n.PitchClass().Int()+iv.Semitones())%12 + 12) % 12
			if got.PitchClass().Int() != want {
				t.Errorf("%s.Transpose(%s) = %s, pitch class %s; want %d", n, iv, got, got.PitchClass(), want)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no transposition was in range")
	}
}

func TestTransposeRoundTrip(t *testing.T) {
	checked := 0
	for _, n := range allNotes() {
		for _, iv := range allIntervals() {
			if iv.Number() > 7 {
				continue
			}
			up, err := n.Transpose(iv)
			if err != nil {
				continue
			}
			checked++
			if back, err := up.TransposeDown(iv); err != nil || back != n {
				t.Errorf("%s.Transpose(%s) = %s, then TransposeDown = %v, %v; want %s", n, iv, up, back, err, n)
			}
			if between, err := material.IntervalBetween(n, up); err != nil || between != iv {
				t.Errorf("IntervalBetween(%s, %s) = %v, %v; want %s", n, up, between, err, iv)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no transposition was in range")
	}
}
