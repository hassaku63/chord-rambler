package performance

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
)

// The twelve major keys (docs/architecture/engine.md, 土台).
var tonics = []string{"C", "G", "D", "A", "E", "B", "F♯", "D♭", "A♭", "E♭", "B♭", "F"}

// kinds is every kind of the vocabulary (docs/architecture/theory-conventions.md,
// 6) and the tensions that 7, maj7 and m7 take, written after a root.
var kinds = []string{
	"", "m", "dim", "aug", "7", "maj7", "m7", "m7(♭5)", "dim7",
	"sus2", "sus4", "7sus4", "add9", "6", "m6", "m(add9)",
	"7(9)", "7(♭9)", "7(♯9)", "7(♯11)", "7(13)", "7(♭13)", "7(♭9,♭13)", "7(9,♯11,13)",
	"maj7(9)", "maj7(♯11)", "maj7(9,♯11,13)", "m7(9)", "m7(11)", "m7(9,11,13)",
}

// basses is no bass, then every simple interval as the bass: members (C/E) and
// notes that are not (C/D, C/B♭).
var basses = []string{"", "P1", "m2", "M2", "m3", "M3", "P4", "A4", "P5", "m6", "M6", "m7", "M7"}

func interval(t *testing.T, s string) material.Interval {
	t.Helper()
	quality := map[byte]material.Quality{
		'P': material.Perfect, 'M': material.Major, 'm': material.Minor,
		'A': material.Augmented, 'd': material.Diminished,
	}[s[0]]
	iv, err := material.NewInterval(int(s[1]-'0'), quality)
	if err != nil {
		t.Fatalf("NewInterval(%s): %v", s, err)
	}
	return iv
}

// eachProgression calls f with a progression in each of the twelve major keys
// and the same progression in C, for every chord of the kinds and the basses on
// every root of at most one sharp or flat. The chord sounds twice, between a
// slot with no chord and one with none chosen, so that no slot starts on a
// multiple of its own length. Only the failure the materials document is
// skipped: a chord beyond a double sharp or a double flat.
func eachProgression(t *testing.T, f func(tonic string, p, inC progression.Progression)) {
	t.Helper()
	compared := 0
	for l := material.C; l <= material.B; l++ {
		for a := -1; a <= 1; a++ {
			root, err := material.NewNoteName(l, a)
			if err != nil {
				t.Fatal(err)
			}
			for _, k := range kinds {
				for _, b := range basses {
					name := notation.FormatNoteName(root) + k
					c, err := notation.ParseChord(name)
					if err == nil && b != "" {
						c, err = c.WithBass(interval(t, b))
					}
					if errors.Is(err, material.ErrAlterOutOfRange) {
						continue
					}
					if err != nil {
						t.Fatalf("%s with the bass %s: %v", name, b, err)
					}
					inC := prog(t,
						slot(t, progression.NoChord(), 1), slot(t, progression.ChordOf(c), 3),
						slot(t, progression.Unchosen(), 2), slot(t, progression.ChordOf(c), 2))
					for _, tonic := range tonics {
						up, err := material.IntervalBetween(note(t, "C"), note(t, tonic))
						if err != nil {
							t.Fatal(err)
						}
						p, err := inC.Transpose(up)
						if errors.Is(err, material.ErrAlterOutOfRange) {
							continue
						}
						if err != nil {
							t.Fatalf("Transpose %s to %s: %v", name, tonic, err)
						}
						compared++
						f(tonic, p, inC)
					}
				}
			}
		}
	}
	if compared == 0 {
		t.Fatal("no progression was checked")
	}
}

func note(t *testing.T, s string) material.NoteName {
	t.Helper()
	n, err := notation.ParseNoteName(s)
	if err != nil {
		t.Fatalf("ParseNoteName(%q): %v", s, err)
	}
	return n
}

// What the notes of a slot hold follows the chord of the slot (誤り 1 in
// docs/architecture/explanations.md): every note is of the pitch classes that
// sound, the upper voices hold every note that decides the chord, the bass is
// the lowest note and the only one below C4, and the upper voices start in C4
// to B4 on the root.
func TestNotesFollowTheChord(t *testing.T) {
	eachProgression(t, func(tonic string, p, _ progression.Progression) {
		notes := Notes(p)
		start := 0
		for _, s := range p.Slots() {
			c, ok := s.Content().Chord()
			if ok {
				checkChord(t, tonic, c, notesAt(notes, start))
			}
			start += s.Beats()
		}
	})
}

func notesAt(notes []Note, start int) []Note {
	var out []Note
	for _, n := range notes {
		if n.Start == start {
			out = append(out, n)
		}
	}
	return out
}

func checkChord(t *testing.T, tonic string, c material.Chord, notes []Note) {
	t.Helper()
	var below, upper []Note
	for _, n := range notes {
		if !c.SoundingPitchClasses().Contains(pitchClass(t, n.Pitch)) {
			t.Errorf("%s major, %v: %d is not of the sounding pitch classes %v", tonic, c.Notes(), n.Pitch, c.SoundingPitchClasses())
		}
		if n.Pitch < 60 {
			below = append(below, n)
		} else {
			upper = append(upper, n)
		}
	}
	bassNote := c.Root()
	if iv, ok := c.Bass(); ok {
		bassNote, _ = c.Root().Transpose(iv)
	}
	if len(below) != 1 || below[0].Pitch < 36 || below[0].Pitch > 47 || pitchClass(t, below[0].Pitch) != bassNote.PitchClass() {
		t.Errorf("%s major, %v: below C4 %v; want the bass %s alone in C2 to B2", tonic, c.Notes(), below, bassNote)
	}
	if len(upper) != len(c.Intervals()) {
		t.Errorf("%s major, %v: %d upper voices; want one for each of the %d members", tonic, c.Notes(), len(upper), len(c.Intervals()))
		return
	}
	lowest := slices.MinFunc(upper, func(a, b Note) int { return a.Pitch - b.Pitch })
	if lowest.Pitch < 60 || lowest.Pitch > 71 || pitchClass(t, lowest.Pitch) != c.Root().PitchClass() {
		t.Errorf("%s major, %v: the lowest upper voice is %d; want the root in C4 to B4", tonic, c.Notes(), lowest.Pitch)
	}
	var upperSet material.PitchClassSet
	for _, n := range upper {
		upperSet = upperSet.Union(material.NewPitchClassSet(pitchClass(t, n.Pitch)))
	}
	if !upperSet.IsSubsetOf(c.PitchClasses()) {
		t.Errorf("%s major, %v: the upper voices %v hold a note that is not a member", tonic, c.Notes(), upperSet)
	}
	for _, d := range deciding(c) {
		if !upperSet.Contains(d) {
			t.Errorf("%s major, %v: the upper voices %v lack %v, which decides the chord", tonic, c.Notes(), upperSet, d)
		}
	}
}

// deciding returns the pitch classes of the notes that decide the chord: the
// third, or without one the second and the fourth (sus), and the seventh.
func deciding(c material.Chord) []material.PitchClass {
	hasThird := slices.ContainsFunc(c.Intervals(), func(iv material.Interval) bool { return iv.Number() == 3 })
	var out []material.PitchClass
	for _, iv := range c.Intervals() {
		n := iv.Number()
		if n == 3 || n == 7 || (!hasThird && (n == 2 || n == 4)) {
			m, _ := c.Root().Transpose(iv) // NewChord checked the members
			out = append(out, m.PitchClass())
		}
	}
	return out
}

func pitchClass(t *testing.T, pitch int) material.PitchClass {
	t.Helper()
	pc, err := material.NewPitchClass(pitch % 12)
	if err != nil {
		t.Fatalf("pitch %d: %v", pitch, err)
	}
	return pc
}

// The notes come in order of their start, each slot with a chord gives notes
// that start where the lengths of the slots before it add up to and last its
// length, and the slots with no chord or none chosen give none. Every note has
// the one velocity, 0 to 127.
func TestNotesFollowTheSlots(t *testing.T) {
	eachProgression(t, func(tonic string, p, _ progression.Progression) {
		notes := Notes(p)
		if !slices.IsSortedFunc(notes, func(a, b Note) int { return a.Start - b.Start }) {
			t.Errorf("%s major: notes out of order: %v", tonic, notes)
		}
		start, count := 0, 0
		for _, s := range p.Slots() {
			at := notesAt(notes, start)
			_, ok := s.Content().Chord()
			if ok != (len(at) > 0) {
				t.Errorf("%s major: the slot at beat %d (a chord: %v) gives %v", tonic, start, ok, at)
			}
			for _, n := range at {
				if n.Beats != s.Beats() || n.Velocity != Velocity {
					t.Errorf("%s major: %+v in a slot of %d beats", tonic, n, s.Beats())
				}
			}
			count += len(at)
			start += s.Beats()
		}
		if count != len(notes) {
			t.Errorf("%s major: %d notes, %d of them at the start of a slot", tonic, len(notes), count)
		}
	})
	if Velocity < 0 || Velocity > 127 {
		t.Errorf("Velocity = %d; want 0 to 127", Velocity)
	}
}

// Transposing the progression moves every note by the semitones of the
// interval, up to an octave where the register wraps around, and keeps the
// starts, the lengths and the velocities.
func TestNotesMoveWithTheKey(t *testing.T) {
	eachProgression(t, func(tonic string, p, inC progression.Progression) {
		up, err := material.IntervalBetween(note(t, "C"), note(t, tonic))
		if err != nil {
			t.Fatal(err)
		}
		before, after := Notes(inC), Notes(p)
		if len(before) != len(after) {
			t.Fatalf("%s major: %d notes; want %d as in C", tonic, len(after), len(before))
		}
		for i, b := range before {
			a := after[i]
			moved := ((a.Pitch-b.Pitch)%12 + 12) % 12
			if moved != up.Semitones()%12 || a.Start != b.Start || a.Beats != b.Beats || a.Velocity != b.Velocity {
				t.Errorf("%s major: note %d is %+v; want %+v moved by %d semitones", tonic, i, a, b, up.Semitones())
			}
		}
	})
}
