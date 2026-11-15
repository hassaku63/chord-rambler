package material_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

func scale(tonic material.NoteName, mode material.Mode) material.Scale {
	s, err := material.NewScale(tonic, mode)
	if err != nil {
		panic(err)
	}
	return s
}

// The twelve major keys (docs/architecture/engine.md) agree with their key
// signatures. The expected notes come from the signature, not from intervals:
// every letter is natural except the first sharps of F C G D A E B, or the
// first flats of B E A D G C F. Source: Hutchinson 2.3 Major Key Signatures
// (https://musictheory.pugetsound.edu/mt21c/MajorKeySignatures.html), the
// order of sharps and flats and Figures 2.3.2 and 2.3.3.
func TestMajorScalesMatchKeySignatures(t *testing.T) {
	tests := []struct {
		tonic      material.NoteName
		accidental int // sharp or flat
		count      int
	}{
		{note(material.C, nat), nat, 0},
		{note(material.G, nat), sharp, 1},
		{note(material.D, nat), sharp, 2},
		{note(material.A, nat), sharp, 3},
		{note(material.E, nat), sharp, 4},
		{note(material.B, nat), sharp, 5},
		{note(material.F, sharp), sharp, 6},
		{note(material.D, flat), flat, 5},
		{note(material.A, flat), flat, 4},
		{note(material.E, flat), flat, 3},
		{note(material.B, flat), flat, 2},
		{note(material.F, nat), flat, 1},
	}
	for _, tt := range tests {
		t.Run(tt.tonic.String(), func(t *testing.T) {
			order := "FCGDAEB"
			if tt.accidental == flat {
				order = "BEADGCF"
			}
			signed := order[:tt.count]
			var want []material.NoteName
			for i := range 7 {
				l := material.Letter((int(tt.tonic.Letter()) + i) % 7)
				alter := nat
				if strings.Contains(signed, l.String()) {
					alter = tt.accidental
				}
				want = append(want, note(l, alter))
			}
			if got := scale(tt.tonic, material.Ionian).Notes(); !slices.Equal(got, want) {
				t.Errorf("%s Ionian = %v, want %v", tt.tonic, got, want)
			}
		})
	}
}

// Source: Hutchinson 31.9 Scales, Table 31.9.4
// (https://musictheory.pugetsound.edu/mt21c/JazzScales.html), the scale
// degrees of Ionian, Dorian, Phrygian, Lydian, Mixolydian, Aeolian and
// Locrian, on C.
func TestModesOnC(t *testing.T) {
	tests := []struct {
		mode  material.Mode
		notes string
	}{
		{material.Ionian, "C D E F G A B"},
		{material.Dorian, "C D E♭ F G A B♭"},
		{material.Phrygian, "C D♭ E♭ F G A♭ B♭"},
		{material.Lydian, "C D E F♯ G A B"},
		{material.Mixolydian, "C D E F G A B♭"},
		{material.Aeolian, "C D E♭ F G A♭ B♭"},
		{material.Locrian, "C D♭ E♭ F G♭ A♭ B♭"},
	}
	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			s := scale(note(material.C, nat), tt.mode)
			if got := s.Notes(); !slices.Equal(got, notes(tt.notes)) {
				t.Errorf("Notes() = %v, want %s", got, tt.notes)
			}
			var want material.PitchClassSet
			for _, n := range notes(tt.notes) {
				want = want.Union(material.NewPitchClassSet(n.PitchClass()))
			}
			if got := s.PitchClasses(); got != want {
				t.Errorf("PitchClasses() = %s, want %s", got, want)
			}
		})
	}
}

// The diatonic triads and seventh chords of C major are in the C major scale.
// Sources: Hutchinson 7.2 Diatonic Chords in Major (the pattern M m m M M m °)
// and 8.3 Roman Numerals of Diatonic Seventh Chords (Figure 8.3.1).
func TestDiatonicChordsOfCMajorAreInTheScale(t *testing.T) {
	cMajor := scale(note(material.C, nat), material.Ionian)
	tests := []struct {
		name    string
		chord   material.Chord
		triad   material.Triad
		seventh material.Seventh
	}{
		{"I", chord(note(material.C, nat), "M3 P5"), material.TriadMajor, material.NoSeventh},
		{"ii", chord(note(material.D, nat), "m3 P5"), material.TriadMinor, material.NoSeventh},
		{"iii", chord(note(material.E, nat), "m3 P5"), material.TriadMinor, material.NoSeventh},
		{"IV", chord(note(material.F, nat), "M3 P5"), material.TriadMajor, material.NoSeventh},
		{"V", chord(note(material.G, nat), "M3 P5"), material.TriadMajor, material.NoSeventh},
		{"vi", chord(note(material.A, nat), "m3 P5"), material.TriadMinor, material.NoSeventh},
		{"vii°", chord(note(material.B, nat), "m3 d5"), material.TriadDiminished, material.NoSeventh},
		{"Imaj7", chord(note(material.C, nat), "M3 P5 M7"), material.TriadMajor, material.SeventhMajor},
		{"ii7", chord(note(material.D, nat), "m3 P5 m7"), material.TriadMinor, material.SeventhMinor},
		{"iii7", chord(note(material.E, nat), "m3 P5 m7"), material.TriadMinor, material.SeventhMinor},
		{"IVmaj7", chord(note(material.F, nat), "M3 P5 M7"), material.TriadMajor, material.SeventhMajor},
		{"V7", chord(note(material.G, nat), "M3 P5 m7"), material.TriadMajor, material.SeventhMinor},
		{"vi7", chord(note(material.A, nat), "m3 P5 m7"), material.TriadMinor, material.SeventhMinor},
		{"viiø7", chord(note(material.B, nat), "m3 d5 m7"), material.TriadDiminished, material.SeventhMinor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.chord.Notes()
			if !cMajor.Contains(tt.chord.PitchClasses()) {
				t.Errorf("%v is not in %v", in, cMajor.Notes())
			}
			if got := tt.chord.Triad(); got != tt.triad {
				t.Errorf("%v: Triad() = %s, want %s", in, got, tt.triad)
			}
			if got := tt.chord.Seventh(); got != tt.seventh {
				t.Errorf("%v: Seventh() = %s, want %s", in, got, tt.seventh)
			}
		})
	}
}

// A♭ major (♭VI, borrowed from the parallel minor) is not in C major but is in
// C Aeolian. Source: Hutchinson 19.1 Mode Mixture
// (https://musictheory.pugetsound.edu/mt21c/ModeMixtureSection.html): ♭6 is
// the most common borrowed degree.
func TestBorrowedChordIsInAeolian(t *testing.T) {
	aFlat := chord(note(material.A, flat), "M3 P5")
	if scale(note(material.C, nat), material.Ionian).Contains(aFlat.PitchClasses()) {
		t.Errorf("A♭ major is in C Ionian")
	}
	if !scale(note(material.C, nat), material.Aeolian).Contains(aFlat.PitchClasses()) {
		t.Errorf("A♭ major is not in C Aeolian")
	}
}

func TestNewScaleRejects(t *testing.T) {
	tests := []struct {
		name  string
		tonic material.NoteName
		mode  material.Mode
		want  error
	}{
		{"mode below Ionian", note(material.C, nat), material.Mode(-1), material.ErrInvalidMode},
		{"mode past Locrian", note(material.C, nat), material.Locrian + 1, material.ErrInvalidMode},
		{"note beyond a double sharp", note(material.G, sharp2), material.Ionian, material.ErrAlterOutOfRange},
		{"note beyond a double flat", note(material.F, flat2), material.Locrian, material.ErrAlterOutOfRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := material.NewScale(tt.tonic, tt.mode); !errors.Is(err, tt.want) {
				t.Errorf("NewScale(%s, %s) error = %v, want %v", tt.tonic, tt.mode, err, tt.want)
			}
		})
	}
	if got := (material.Scale{}).Notes(); !slices.Equal(got, notes("C D E F G A B")) {
		t.Errorf("zero Scale Notes() = %v, want C major", got)
	}
}
