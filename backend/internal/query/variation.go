package query

import (
	"errors"
	"fmt"
	"slices"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// ErrNoTriadQuality reports a chord whose triad is not major, minor or
// diminished: a sus chord, a power chord, an augmented triad or any other
// shape (material.TriadOther). The list of variations follows the quality of
// the triad, so such a chord has none.
var ErrNoTriadQuality = errors.New("chord has no major, minor or diminished triad")

// Variation is one variation of a chord: a chord on the same root, without a
// bass, whose sixth, seventh, sus or tension differs (変形 in
// docs/glossary.md), and its labels.
type Variation struct {
	// Chord is the variation.
	Chord material.Chord
	// IsInput reports whether the variation is the chord asked about itself,
	// bass included.
	IsInput  bool
	Members  MembersLabel
	Reading  ReadingLabel
	Cautions CautionsLabel
	OutOfKey OutOfKeyLabel
}

// MembersLabel holds the members of the variation, the root first
// (material.Chord.Notes).
type MembersLabel struct {
	Notes []material.NoteName
}

// Basis returns Definition.
func (MembersLabel) Basis() Basis { return Definition }

// ReadingLabel holds the main reading of the variation in the key
// (reading.ReadChord): the kind, the function, the target and the modes.
type ReadingLabel struct {
	Main reading.Reading
}

// Basis returns School.
func (ReadingLabel) Basis() Basis { return School }

// CautionsLabel holds the tensions of the variation that tend to sound muddy
// in the key (reading.TensionCautions).
type CautionsLabel struct {
	Cautions []reading.TensionCaution
}

// Basis returns School.
func (CautionsLabel) Basis() Basis { return School }

// OutOfKeyLabel holds the notes of the variation that are outside the key:
// members whose spelling is not among the notes of the scale of the key (the
// same comparison as the kind of a reading, #102), in the order of
// material.Chord.Notes. No notes means it has none. The chord asked about is
// not looked at.
type OutOfKeyLabel struct {
	Notes []material.NoteName
}

// Basis returns Computation.
func (OutOfKeyLabel) Basis() Basis { return Computation }

// variationShapes is the fixed list of variations for each quality of triad,
// in order, as intervals above the root (the owner's list on #115). The
// shapes follow the MusicXML 4.0 kind values
// (https://www.w3.org/2021/06/musicxml40/musicxml-reference/data-types/kind-value/)
// and Harte et al., ISMIR 2005, table 2
// (https://archives.ismir.net/ismir2005/paper/000080.pdf): 6 and m6 add a
// major sixth (maj6 (3,5,6) and min6 (b3,5,6) in Harte), 9 is a dominant
// seventh with a major ninth, m9 a minor seventh with a major ninth, and sus2
// and sus4 keep the perfect fifth. 7sus4 is sus4 with a minor seventh, and
// add9 and m(add9) add a major ninth without a seventh
// (docs/architecture/theory-conventions.md, 6).
var variationShapes = map[material.Triad][][]material.Interval{
	material.TriadMajor: {
		{maj3, p5},             // triad
		{maj3, p5, maj7},       // maj7
		{maj3, p5, min7},       // 7
		{maj3, p5, maj6},       // 6
		{maj3, p5, maj9},       // add9
		{maj2, p5},             // sus2
		{p4, p5},               // sus4
		{p4, p5, min7},         // 7sus4
		{maj3, p5, min7, maj9}, // 9
	},
	material.TriadMinor: {
		{min3, p5},             // triad
		{min3, p5, min7},       // m7
		{min3, p5, maj6},       // m6
		{min3, p5, maj9},       // m(add9)
		{min3, p5, min7, maj9}, // m9
		{maj2, p5},             // sus2
		{p4, p5},               // sus4
	},
	material.TriadDiminished: {
		{min3, dim5},       // triad
		{min3, dim5, min7}, // m7♭5
		{min3, dim5, dim7}, // dim7
	},
}

var (
	maj2 = interval(2, material.Major)
	min3 = interval(3, material.Minor)
	maj3 = interval(3, material.Major)
	p4   = interval(4, material.Perfect)
	dim5 = interval(5, material.Diminished)
	p5   = interval(5, material.Perfect)
	maj6 = interval(6, material.Major)
	dim7 = interval(7, material.Diminished)
	min7 = interval(7, material.Minor)
	maj7 = interval(7, material.Major)
	maj9 = interval(9, material.Major)
)

func interval(number int, q material.Quality) material.Interval {
	iv, err := material.NewInterval(number, q)
	if err != nil {
		panic(err) // the table above names valid intervals only
	}
	return iv
}

// Variations returns the variations of the chord in the major key, in the
// order of the fixed list for the quality of its triad (material.Chord.Triad;
// a bass is not looked at, so C/E has the list of C):
//
//   - major: the triad, maj7, 7, 6, add9, sus2, sus4, 7sus4, 9;
//   - minor: the triad, m7, m6, m(add9), m9, sus2, sus4;
//   - diminished: the triad, m7♭5, dim7.
//
// Every variation is on the root of the chord and has no bass. The one equal
// to the chord itself, bass included, is marked (IsInput), so C/E marks none.
// Each carries its labels; a variation that has notes from outside the key is
// listed as well.
//
// It fails for a key other than Ionian (relation.ErrNotMajor), for a chord
// whose triad is not major, minor or diminished (ErrNoTriadQuality), and when
// a variation would need more than a double sharp or a double flat
// (material.ErrAlterOutOfRange; dim7 on C♭).
func Variations(c material.Chord, key material.Scale) ([]Variation, error) {
	if key.Mode() != material.Ionian {
		return nil, fmt.Errorf("variations in %s %s: %w", key.Tonic(), key.Mode(), relation.ErrNotMajor)
	}
	shapes, ok := variationShapes[c.Triad()]
	if !ok {
		return nil, fmt.Errorf("variations of the chord on %s (%s): %w", c.Root(), c.Triad(), ErrNoTriadQuality)
	}
	out := make([]Variation, 0, len(shapes))
	for _, shape := range shapes {
		v, err := material.NewChord(c.Root(), shape...)
		if err != nil {
			return nil, fmt.Errorf("variations of the chord on %s: %w", c.Root(), err)
		}
		read, err := reading.ReadChord(v, key)
		if err != nil {
			return nil, fmt.Errorf("variations: %w", err)
		}
		cautions, err := reading.TensionCautions(v, key)
		if err != nil {
			return nil, fmt.Errorf("variations: %w", err)
		}
		out = append(out, Variation{
			Chord:    v,
			IsInput:  v == c,
			Members:  MembersLabel{Notes: v.Notes()},
			Reading:  ReadingLabel{Main: read.Main},
			Cautions: CautionsLabel{Cautions: cautions},
			OutOfKey: OutOfKeyLabel{Notes: notesOutOfKey(v, key)},
		})
	}
	return out, nil
}

// notesOutOfKey returns the members of the variation whose spelling is not
// among the notes of the scale of the key.
func notesOutOfKey(v material.Chord, key material.Scale) []material.NoteName {
	inKey := key.Notes()
	var out []material.NoteName
	for _, n := range v.Notes() {
		if !slices.Contains(inKey, n) {
			out = append(out, n)
		}
	}
	return out
}
