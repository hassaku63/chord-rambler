// The tests sit in the package itself, since the engine's import rules
// (backend/.golangci.yml) do not list performance, but use only its exported
// names.

package performance

import (
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

func parse(t *testing.T, name string) progression.Content {
	t.Helper()
	c, err := notation.ParseChord(name)
	if err != nil {
		t.Fatalf("ParseChord(%q): %v", name, err)
	}
	return progression.ChordOf(c)
}

// prog returns the progression in C major of the slots in order.
func prog(t *testing.T, slots ...progression.Slot) progression.Progression {
	t.Helper()
	p := progression.Progression{}
	for _, s := range slots {
		var err error
		if p, err = p.Append(s); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	return p
}

func slot(t *testing.T, c progression.Content, beats int) progression.Slot {
	t.Helper()
	s, err := progression.NewSlot(c, beats)
	if err != nil {
		t.Fatalf("NewSlot(%d): %v", beats, err)
	}
	return s
}

// at returns the notes of one slot: the note numbers, all starting on the beat
// and lasting the beats, at the one velocity.
func at(start, beats int, pitches ...int) []Note {
	notes := make([]Note, 0, len(pitches))
	for _, p := range pitches {
		notes = append(notes, Note{Pitch: p, Start: start, Beats: beats, Velocity: Velocity})
	}
	return notes
}

// The note numbers are MIDI's with C4 = 60, so C2 = 36. Each slot sounds the
// bass in C2 to B2, then the root in C4 to B4 and the members above it by the
// semitones of their intervals.
func TestNotes(t *testing.T) {
	p := prog(t,
		slot(t, parse(t, "C"), 4),          // beat 0
		slot(t, parse(t, "G7"), 4),         // beat 4
		slot(t, progression.NoChord(), 3),  // beat 8: nothing sounds
		slot(t, parse(t, "Dm7/C"), 1),      // beat 11
		slot(t, parse(t, "C6"), 2),         // beat 12
		slot(t, progression.Unchosen(), 2), // beat 14: nothing sounds
		slot(t, parse(t, "Csus4"), 2),      // beat 16
		slot(t, parse(t, "G7(♭9)"), 6),     // beat 18
	)
	want := slices.Concat(
		at(0, 4, 36, 60, 64, 67),          // C2; C4 E4 G4
		at(4, 4, 43, 67, 71, 74, 77),      // G2; G4 B4 D5 F5
		at(11, 1, 36, 62, 65, 69, 72),     // C2 (the bass, also the seventh); D4 F4 A4 C5
		at(12, 2, 36, 60, 64, 67, 69),     // C2; C4 E4 G4 A4
		at(16, 2, 36, 60, 65, 67),         // C2; C4 F4 G4 (no third)
		at(18, 6, 43, 67, 71, 74, 77, 80), // G2; G4 B4 D5 F5 A♭5
	)
	if got := Notes(p); !slices.Equal(got, want) {
		t.Errorf("Notes:\n got %v\nwant %v", got, want)
	}
}

// Notes are placed by pitch, not by spelling, and a bass that is not a member
// sounds only as the bass.
func TestNotesOfOneChord(t *testing.T) {
	tests := []struct {
		name string
		want []int
	}{
		{"B♯", []int{36, 60, 64, 67}},                              // B♯ D𝄪 F𝄪 sound as C E G
		{"C♭", []int{47, 71, 75, 78}},                              // C♭ E♭ G♭ sound as B D♯ F♯
		{"B", []int{47, 71, 75, 78}},                               // the highest root
		{"C/E", []int{40, 60, 64, 67}},                             // the bass is a member and sounds above too
		{"C/D", []int{38, 60, 64, 67}},                             // D is not a member: the bass only
		{"C/B♭", []int{46, 60, 64, 67}},                            // B♭ is not a member: the bass only
		{"F♯m7(♭5)", []int{42, 66, 69, 72, 76}},                    // F♯2; F♯4 A4 C5 E5
		{"Bmaj7(9,♯11,13)", []int{47, 71, 75, 78, 82, 85, 89, 92}}, // B2; B4 D♯5 F♯5 A♯5 C♯6 E♯6 G♯6
	}
	for _, tt := range tests {
		p := prog(t, slot(t, parse(t, tt.name), 4))
		if got := Notes(p); !slices.Equal(got, at(0, 4, tt.want...)) {
			t.Errorf("Notes(%s) = %v; want %v", tt.name, got, tt.want)
		}
	}
}

// Slots with no chord or none chosen give no notes.
func TestNotesOfSilentSlots(t *testing.T) {
	tests := []struct {
		name string
		p    progression.Progression
	}{
		{"no slots", progression.Progression{}},
		{"no chord", prog(t, slot(t, progression.NoChord(), 4))},
		{"none chosen", prog(t, slot(t, progression.Unchosen(), 4))},
	}
	for _, tt := range tests {
		if got := Notes(tt.p); len(got) != 0 {
			t.Errorf("%s: Notes = %v; want none", tt.name, got)
		}
	}
}
