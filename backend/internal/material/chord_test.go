package material_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// chord builds a chord for a test table, like note.
func chord(root material.NoteName, intervals string) material.Chord {
	c, err := material.NewChord(root, ivs(intervals)...)
	if err != nil {
		panic(err)
	}
	return c
}

// Sources (Hutchinson, Music Theory for the 21st-Century Classroom,
// https://musictheory.pugetsound.edu/mt21c/):
//   - 6.2 Lead-Sheet Symbols: the F, Gm, D° and C+ rows.
//   - 6.5 Simple "Sus" Chords and 6.6 Summary: sus2 and sus4 (the second or
//     fourth replaces the third; a perfect fifth).
//   - 8.3 Roman Numerals of Diatonic Seventh Chords: 7 is a minor seventh, M7
//     a major one; ø7 is a diminished triad with a minor seventh, °7 one with
//     a diminished seventh. Figure 8.3.1 spells V7, IM7, ii7 and viiø7 in C.
//   - 31.1 Jazz Chord Basics: a 9 has the seventh below it, a ninth with no
//     seventh is add 9, maj applies to the seventh and m to the third; C9(13)
//     is C–E–G–B♭–D–A.
//   - 31.2 Chord Symbol Specifics: C7sus is C–F–G–B♭; ♯11 over a normal fifth.
//   - 31.3 Altered Dominant Seventh Chords: ♭9, ♯9, and ♭13 with the fifth.
//
// MusicXML 4.0, kind-value (https://www.w3.org/2021/06/musicxml40/musicxml-reference/data-types/kind-value/):
// minor-11th (the minor-ninth chord and a perfect 11th) and power (a perfect fifth).
func TestChordVocabulary(t *testing.T) {
	ts := func(t ...material.Tension) []material.Tension { return t }
	tests := []struct {
		name      string
		root      material.NoteName
		intervals string
		notes     string
		pcs       []int
		triad     material.Triad
		seventh   material.Seventh
		tensions  []material.Tension
	}{
		{"F", note(material.F, nat), "M3 P5", "F A C", []int{0, 5, 9}, material.TriadMajor, material.NoSeventh, nil},
		{"Gm", note(material.G, nat), "m3 P5", "G B♭ D", []int{2, 7, 10}, material.TriadMinor, material.NoSeventh, nil},
		{"D°", note(material.D, nat), "m3 d5", "D F A♭", []int{2, 5, 8}, material.TriadDiminished, material.NoSeventh, nil},
		{"C+", note(material.C, nat), "M3 A5", "C E G♯", []int{0, 4, 8}, material.TriadAugmented, material.NoSeventh, nil},
		{"F♯", note(material.F, sharp), "M3 P5", "F♯ A♯ C♯", []int{1, 6, 10}, material.TriadMajor, material.NoSeventh, nil},
		{"E♭m", note(material.E, flat), "m3 P5", "E♭ G♭ B♭", []int{3, 6, 10}, material.TriadMinor, material.NoSeventh, nil},
		{"G7", note(material.G, nat), "M3 P5 m7", "G B D F", []int{2, 5, 7, 11}, material.TriadMajor, material.SeventhMinor, nil},
		{"Cmaj7", note(material.C, nat), "M3 P5 M7", "C E G B", []int{0, 4, 7, 11}, material.TriadMajor, material.SeventhMajor, nil},
		{"Dm7", note(material.D, nat), "m3 P5 m7", "D F A C", []int{0, 2, 5, 9}, material.TriadMinor, material.SeventhMinor, nil},
		{"Bm7♭5", note(material.B, nat), "m3 d5 m7", "B D F A", []int{2, 5, 9, 11}, material.TriadDiminished, material.SeventhMinor, nil},
		{"Cdim7", note(material.C, nat), "m3 d5 d7", "C E♭ G♭ B♭♭", []int{0, 3, 6, 9}, material.TriadDiminished, material.SeventhDiminished, nil},
		{"Csus2", note(material.C, nat), "M2 P5", "C D G", []int{0, 2, 7}, material.TriadSus2, material.NoSeventh, nil},
		{"Csus4", note(material.C, nat), "P4 P5", "C F G", []int{0, 5, 7}, material.TriadSus4, material.NoSeventh, nil},
		{"C7sus4", note(material.C, nat), "P4 P5 m7", "C F G B♭", []int{0, 5, 7, 10}, material.TriadSus4, material.SeventhMinor, nil},
		{"Cadd9", note(material.C, nat), "M3 P5 M9", "C E G D", []int{0, 2, 4, 7}, material.TriadMajor, material.NoSeventh, ts(material.Tension9)},
		{"C9", note(material.C, nat), "M3 P5 m7 M9", "C E G B♭ D", []int{0, 2, 4, 7, 10}, material.TriadMajor, material.SeventhMinor, ts(material.Tension9)},
		{"Cmaj9", note(material.C, nat), "M3 P5 M7 M9", "C E G B D", []int{0, 2, 4, 7, 11}, material.TriadMajor, material.SeventhMajor, ts(material.Tension9)},
		{"Cm9", note(material.C, nat), "m3 P5 m7 M9", "C E♭ G B♭ D", []int{0, 2, 3, 7, 10}, material.TriadMinor, material.SeventhMinor, ts(material.Tension9)},
		{"C7(♭9)", note(material.C, nat), "M3 P5 m7 m9", "C E G B♭ D♭", []int{0, 1, 4, 7, 10}, material.TriadMajor, material.SeventhMinor, ts(material.TensionFlat9)},
		{"C7(♯9)", note(material.C, nat), "M3 P5 m7 A9", "C E G B♭ D♯", []int{0, 3, 4, 7, 10}, material.TriadMajor, material.SeventhMinor, ts(material.TensionSharp9)},
		{"C7(♯11)", note(material.C, nat), "M3 P5 m7 A11", "C E G B♭ F♯", []int{0, 4, 6, 7, 10}, material.TriadMajor, material.SeventhMinor, ts(material.TensionSharp11)},
		{"C7(♭13)", note(material.C, nat), "M3 P5 m7 m13", "C E G B♭ A♭", []int{0, 4, 7, 8, 10}, material.TriadMajor, material.SeventhMinor, ts(material.TensionFlat13)},
		{"C9(13)", note(material.C, nat), "M3 P5 m7 M9 M13", "C E G B♭ D A", []int{0, 2, 4, 7, 9, 10}, material.TriadMajor, material.SeventhMinor, ts(material.Tension9, material.Tension13)},
		{"Cm11", note(material.C, nat), "m3 P5 m7 M9 P11", "C E♭ G B♭ D F", []int{0, 2, 3, 5, 7, 10}, material.TriadMinor, material.SeventhMinor, ts(material.Tension9, material.Tension11)},
		{"C5", note(material.C, nat), "P5", "C G", []int{0, 7}, material.TriadOther, material.NoSeventh, nil},
		// The project's rule: a second and a fourth with no third is neither sus.
		{"C2 and 4", note(material.C, nat), "M2 P4 P5", "C D F G", []int{0, 2, 5, 7}, material.TriadOther, material.NoSeventh, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chord(tt.root, tt.intervals)
			in := tt.root.String() + " " + tt.intervals
			if got := c.Notes(); !slices.Equal(got, notes(tt.notes)) {
				t.Errorf("chord on %s: Notes() = %v, want %s", in, got, tt.notes)
			}
			if got := ints(c.PitchClasses()); !slices.Equal(got, tt.pcs) {
				t.Errorf("chord on %s: PitchClasses() = %v, want %v", in, got, tt.pcs)
			}
			if got := c.Triad(); got != tt.triad {
				t.Errorf("chord on %s: Triad() = %s, want %s", in, got, tt.triad)
			}
			if got := c.Seventh(); got != tt.seventh {
				t.Errorf("chord on %s: Seventh() = %s, want %s", in, got, tt.seventh)
			}
			if got := c.Tensions(); !slices.Equal(got, tt.tensions) {
				t.Errorf("chord on %s: Tensions() = %v, want %v", in, got, tt.tensions)
			}
		})
	}
}

func TestChordViewsOutsideTheVocabulary(t *testing.T) {
	tests := []struct {
		name      string
		intervals string
		triad     material.Triad
		seventh   material.Seventh
		tensions  []material.Tension
	}{
		{"root alone (the zero value)", "", material.TriadOther, material.NoSeventh, nil},
		{"third without a fifth", "M3 m7", material.TriadOther, material.SeventhMinor, nil},
		{"major third, diminished fifth", "M3 d5", material.TriadOther, material.NoSeventh, nil},
		{"augmented fourth for sus", "A4 P5", material.TriadOther, material.NoSeventh, nil},
		{"augmented seventh", "M3 P5 A7", material.TriadMajor, material.SeventhOther, nil},
		{"diminished eleventh", "M3 P5 m7 d11", material.TriadMajor, material.SeventhMinor, []material.Tension{material.TensionOther}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chord(note(material.C, nat), tt.intervals)
			if got := c.Triad(); got != tt.triad {
				t.Errorf("C %s: Triad() = %s, want %s", tt.intervals, got, tt.triad)
			}
			if got := c.Seventh(); got != tt.seventh {
				t.Errorf("C %s: Seventh() = %s, want %s", tt.intervals, got, tt.seventh)
			}
			if got := c.Tensions(); !slices.Equal(got, tt.tensions) {
				t.Errorf("C %s: Tensions() = %v, want %v", tt.intervals, got, tt.tensions)
			}
		})
	}
	if got := (material.Chord{}).Notes(); !slices.Equal(got, notes("C")) {
		t.Errorf("zero Chord Notes() = %v, want [C]", got)
	}
}

func TestNewChordRejects(t *testing.T) {
	tests := []struct {
		name      string
		root      material.NoteName
		intervals string
		want      error
	}{
		{"major and minor third", note(material.C, nat), "M3 m3 P5", material.ErrDuplicateNumber},
		{"the same interval twice", note(material.C, nat), "M3 P5 P5", material.ErrDuplicateNumber},
		{"a unison besides the root", note(material.C, nat), "P1 M3 P5", material.ErrDuplicateNumber},
		{"a member beyond a double sharp", note(material.B, sharp2), "M3 P5", material.ErrAlterOutOfRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := material.NewChord(tt.root, ivs(tt.intervals)...); !errors.Is(err, tt.want) {
				t.Errorf("NewChord(%s, %s) = %v, %v; want error %v", tt.root, tt.intervals, got.Notes(), err, tt.want)
			}
		})
	}
	if _, err := chord(note(material.D, sharp2), "P5").WithBass(interval(3, material.Major)); !errors.Is(err, material.ErrAlterOutOfRange) {
		t.Errorf("D♯♯ chord WithBass(M3) error = %v, want %v", err, material.ErrAlterOutOfRange)
	}
}

func TestChordBass(t *testing.T) {
	c := chord(note(material.C, nat), "M3 P5")
	if _, ok := c.Bass(); ok {
		t.Errorf("C major without a bass: Bass() reports one")
	}
	tests := []struct {
		name string
		bass material.Interval
	}{
		{"a member (C/E)", interval(3, material.Major)},
		{"not a member (C/D)", interval(2, material.Major)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slash, err := c.WithBass(tt.bass)
			if err != nil {
				t.Fatalf("WithBass(%s): %v", tt.bass, err)
			}
			if got, ok := slash.Bass(); !ok || got != tt.bass {
				t.Errorf("Bass() = %s, %t; want %s, true", got, ok, tt.bass)
			}
			// The bass does not change the members.
			if !slices.Equal(slash.Notes(), c.Notes()) || slash.PitchClasses() != c.PitchClasses() {
				t.Errorf("with bass %s: Notes() = %v, PitchClasses() = %s; want %v, %s",
					tt.bass, slash.Notes(), slash.PitchClasses(), c.Notes(), c.PitchClasses())
			}
		})
	}
}

func TestChordTranspose(t *testing.T) {
	tests := []struct {
		name  string
		chord material.Chord
		by    material.Interval
		notes string
	}{
		{"C up an augmented fourth is F♯", chord(note(material.C, nat), "M3 P5"), interval(4, material.Augmented), "F♯ A♯ C♯"},
		{"Cm up a minor third is E♭m", chord(note(material.C, nat), "m3 P5"), interval(3, material.Minor), "E♭ G♭ B♭"},
		{"E♭m up a minor second is F♭m", chord(note(material.E, flat), "m3 P5"), interval(2, material.Minor), "F♭ A♭♭ C♭"},
		{"G7 up a major second is A7", chord(note(material.G, nat), "M3 P5 m7"), interval(2, material.Major), "A C♯ E G"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.chord.Transpose(tt.by)
			if err != nil {
				t.Fatalf("%v: %v", tt.by, err)
			}
			if !slices.Equal(got.Notes(), notes(tt.notes)) {
				t.Errorf("Notes() = %v, want %s", got.Notes(), tt.notes)
			}
		})
	}

	slash, err := chord(note(material.C, nat), "M3 P5").WithBass(interval(3, material.Major))
	if err != nil {
		t.Fatal(err)
	}
	moved, err := slash.Transpose(interval(2, material.Major))
	if err != nil {
		t.Fatal(err)
	}
	if bass, ok := moved.Bass(); !ok || bass != interval(3, material.Major) || moved.Root() != note(material.D, nat) {
		t.Errorf("C/E up a major second: root %s, bass %s, %t; want D, M3, true", moved.Root(), bass, ok)
	}

	if _, err := chord(note(material.B, sharp), "M3 P5").Transpose(interval(2, material.Augmented)); !errors.Is(err, material.ErrAlterOutOfRange) {
		t.Errorf("B♯ major up an augmented second: error %v, want %v", err, material.ErrAlterOutOfRange)
	}
}

// Every chord of the vocabulary, on every root, moved by every interval: the
// pitch classes turn by the semitones of the interval, and the intervals stay.
// Only results beyond a double sharp or a double flat are skipped.
func TestChordTransposeTurnsPitchClasses(t *testing.T) {
	shapes := []string{
		"M3 P5", "m3 P5", "m3 d5", "M3 A5", "M3 P5 m7", "M3 P5 M7", "m3 P5 m7", "m3 d5 m7", "m3 d5 d7",
		"M2 P5", "P4 P5", "P4 P5 m7", "M3 P5 M9", "M3 P5 m7 M9", "M3 P5 M7 M9", "m3 P5 m7 M9",
		"M3 P5 m7 m9", "M3 P5 m7 A11", "M3 P5 m7 M9 M13", "m3 P5 m7 M9 P11",
	}
	third := interval(3, material.Major)
	checked := 0
	for _, root := range allNotes() {
		for _, shape := range shapes {
			c, err := material.NewChord(root, ivs(shape)...)
			if err == nil && shape == "M3 P5" {
				c, err = c.WithBass(third) // a first inversion, so the bass is moved too
			}
			if err != nil {
				if !errors.Is(err, material.ErrAlterOutOfRange) {
					t.Errorf("NewChord(%s, %s): unexpected error %v", root, shape, err)
				}
				continue
			}
			for _, iv := range allIntervals() {
				moved, err := c.Transpose(iv)
				if err != nil {
					if !errors.Is(err, material.ErrAlterOutOfRange) {
						t.Errorf("chord on %s (%s).Transpose(%s): unexpected error %v", root, shape, iv, err)
					}
					continue
				}
				checked++
				if got, want := moved.PitchClasses(), c.PitchClasses().Transpose(iv.Semitones()); got != want {
					t.Errorf("chord on %s (%s).Transpose(%s): pitch classes %s, want %s", root, shape, iv, got, want)
				}
				if !slices.Equal(moved.Intervals(), c.Intervals()) {
					t.Errorf("chord on %s (%s).Transpose(%s): intervals %v, want %v", root, shape, iv, moved.Intervals(), c.Intervals())
				}
				got, ok := moved.Bass()
				if want, wantOK := c.Bass(); got != want || ok != wantOK {
					t.Errorf("chord on %s (%s).Transpose(%s): bass %s, %t; want %s, %t", root, shape, iv, got, ok, want, wantOK)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no transposition was in range")
	}
}
