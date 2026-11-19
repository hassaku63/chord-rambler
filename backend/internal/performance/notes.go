package performance

import (
	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

// Velocity is the velocity of every note, 0 to 127. There is one value until
// the progression has accompaniment patterns. 64 is the value the MIDI 1.0
// Detailed Specification gives for mezzo-forte and for devices without
// velocity sensitivity.
const Velocity = 64

// The lowest note numbers of the two registers: the bass in C2 to B2 and the
// root of the upper voices in C4 to B4 (C4 = 60).
const (
	bassC2  = 36
	upperC4 = 60
)

// Note is one note to play: a MIDI note number (C4 = 60), the beat it starts
// on counted from 0 at the start of the progression, its length in beats, and
// its velocity (0 to 127).
type Note struct {
	Pitch    int
	Start    int
	Beats    int
	Velocity int
}

// Notes returns the notes of the progression as block chords, in order of
// their start. Each slot with a chord sounds for its whole length: first the
// bass, the bass of the chord or else its root, in C2 to B2; then the upper
// voices, the root in C4 to B4 and each member above it by the semitones of
// its interval, in the order of the chord's intervals. Notes are placed by
// pitch, not by spelling, so B♯ sits where C does. A bass that is not a member
// sounds only as the bass. Slots with no chord or none chosen give no notes,
// and the slots after them start where they would anyway.
func Notes(p progression.Progression) []Note {
	var notes []Note
	start := 0
	for _, s := range p.Slots() {
		if c, ok := s.Content().Chord(); ok {
			for _, pitch := range voice(c) {
				notes = append(notes, Note{Pitch: pitch, Start: start, Beats: s.Beats(), Velocity: Velocity})
			}
		}
		start += s.Beats()
	}
	return notes
}

// voice returns the note numbers of the chord: the bass, then the upper
// voices from the root.
func voice(c material.Chord) []int {
	bass := c.Root()
	if iv, ok := c.Bass(); ok {
		// WithBass checked that the bass is in range.
		bass, _ = c.Root().Transpose(iv)
	}
	root := upperC4 + c.Root().PitchClass().Int()
	ivs := c.Intervals()
	pitches := make([]int, 0, 1+len(ivs))
	pitches = append(pitches, bassC2+bass.PitchClass().Int())
	for _, iv := range ivs {
		pitches = append(pitches, root+iv.Semitones())
	}
	return pitches
}
