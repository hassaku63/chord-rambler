package reading_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// caution is a TensionCaution from the intervals as the tables write them.
func caution(tension, tone string) reading.TensionCaution {
	return reading.TensionCaution{Tension: ivs(tension)[0], ChordTone: ivs(tone)[0]}
}

// The tensions that tend to sound muddy in C major
// (docs/architecture/theory-conventions.md, 5; SoundQuest,
// https://soundquest.jp/quest/chord/chord-mv4/tension-chord-avoid/2/ and
// https://soundquest.jp/quest/chord/chord-mv6/cst-major-chordscales/): a note
// of the scale a half step above the root, the third, the fifth or the
// seventh. On a dominant seventh only the 11 is a caution, so the ♭9 of E7 and
// the ♭13 of A7 (both F, a half step above E) are not. The 13 of Dm7 (B) is a
// whole step above the fifth (A). The notes of the scale are compared by
// spelling, as the kind of a reading is (#102): the ♭9 of E♭maj7 is F♭, not
// the E of the scale.
func TestTensionCautionsInCMajor(t *testing.T) {
	tests := []struct {
		name  string
		chord material.Chord
		want  []reading.TensionCaution
	}{
		// The cases of #103.
		{"G7(11): C over B", chord(gNat, "M3 P5 m7 P11", ""), []reading.TensionCaution{caution("P11", "M3")}},
		{"Cmaj7(9)", chord(cNat, "M3 P5 M7 M9", ""), nil},
		{"Dm7(13)", chord(dNat, "m3 P5 m7 M13", ""), nil},
		// What follows from the rule.
		{"Cmaj7(11): F over E", chord(cNat, "M3 P5 M7 P11", ""), []reading.TensionCaution{caution("P11", "M3")}},
		{
			"Em7(♭9,♭13): F over E, C over B", chord(eNat, "m3 P5 m7 m9 m13", ""),
			[]reading.TensionCaution{caution("m9", "P1"), caution("m13", "P5")},
		},
		{"Am7(11): D is a whole step above C", chord(aNat, "m3 P5 m7 P11", ""), nil},
		{"A7(11): D over C♯", chord(aNat, "M3 P5 m7 P11", ""), []reading.TensionCaution{caution("P11", "M3")}},
		{"E7(♭9): only the 11 on a dominant seventh", chord(eNat, "M3 P5 m7 m9", ""), nil},
		{"A7(♭13): only the 11 on a dominant seventh", chord(aNat, "M3 P5 m7 m13", ""), nil},
		{"G7(♭9): A♭ is not in the scale", chord(gNat, "M3 P5 m7 m9", ""), nil},
		{"E♭maj7(♭9): F♭ is not spelled as a note of the scale", chord(eFlat, "M3 P5 M7 m9", ""), nil},
		{"C: no tensions", chord(cNat, maj, ""), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reading.TensionCautions(tt.chord, major(cNat))
			if err != nil {
				t.Fatalf("TensionCautions: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("TensionCautions = %v; want %v", got, tt.want)
			}
		})
	}
}

// A key other than Ionian is refused, as ReadChord refuses it.
func TestTensionCautionsOtherThanMajor(t *testing.T) {
	for m := material.Dorian; m <= material.Locrian; m++ {
		key, err := material.NewScale(cNat, m)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := reading.TensionCautions(chord(gNat, "M3 P5 m7 P11", ""), key); !errors.Is(err, relation.ErrNotMajor) {
			t.Errorf("TensionCautions in C %s: err = %v; want ErrNotMajor", m, err)
		}
	}
}
