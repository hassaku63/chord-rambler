package reading

import (
	"fmt"
	"slices"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// TensionCaution is a tension that tends to sound muddy on its chord
// (docs/architecture/theory-conventions.md, 5). It is a caution, not a rule:
// nothing is refused.
type TensionCaution struct {
	// Tension is the tension, as an interval above the root (P11 for the 11
	// of G7(11)).
	Tension material.Interval
	// ChordTone is the note of the chord a half step below the tension, as an
	// interval above the root (M3, the B of G7).
	ChordTone material.Interval
}

// TensionCautions returns the tensions (9, 11 and 13) of the chord that tend
// to sound muddy in the major key, from the lowest: those that are notes of
// the scale of the key, by spelling, and a half step above a note of the
// chord (the root, the third, the fifth or the seventh), by pitch class. On a
// dominant seventh (a major triad and a minor seventh, on any degree) only an
// 11 is kept. Nothing else is looked at, so the 13 of IIm7 (B on Dm7 in C) is
// not a caution: it is a whole step above the fifth.
//
// It fails for a key other than Ionian (relation.ErrNotMajor).
func TensionCautions(c material.Chord, key material.Scale) ([]TensionCaution, error) {
	if key.Mode() != material.Ionian {
		return nil, fmt.Errorf("tension cautions in %s %s: %w", key.Tonic(), key.Mode(), relation.ErrNotMajor)
	}
	dominant := c.Triad() == material.TriadMajor && c.Seventh() == material.SeventhMinor
	scale := key.Notes()
	var tones []material.Interval
	for _, iv := range c.Intervals() {
		if n := iv.Number(); n == 1 || n == 3 || n == 5 || n == 7 {
			tones = append(tones, iv)
		}
	}
	notes := c.Notes() // in the order of Intervals
	var out []TensionCaution
	for i, iv := range c.Intervals() {
		if iv.Number() < 9 || (dominant && iv.Number() != 11) {
			continue
		}
		if !slices.Contains(scale, notes[i]) {
			continue
		}
		for _, tone := range tones {
			if ((iv.Semitones()-tone.Semitones())%12+12)%12 == 1 {
				out = append(out, TensionCaution{Tension: iv, ChordTone: tone})
				break
			}
		}
	}
	return out, nil
}
