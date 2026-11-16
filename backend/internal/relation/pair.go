package relation

import (
	"fmt"
	"slices"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// RootMotion returns the simple interval up from the root of the previous
// chord to the root of the next (material.IntervalBetween): G to C is a
// perfect fourth, C to A a major sixth. The basses are not looked at. Saying
// "down a fifth" belongs to the readings and the text. It fails when the two
// roots have no interval of the five qualities (C♯ to E𝄫,
// material.ErrInvalidInterval).
func RootMotion(prev, next material.Chord) (material.Interval, error) {
	iv, err := material.IntervalBetween(prev.Root(), next.Root())
	if err != nil {
		return material.Interval{}, fmt.Errorf("root motion: %w", err)
	}
	return iv, nil
}

// CommonTones returns the pitch classes that sound in both chords (the
// members and the bass, Chord.SoundingPitchClasses).
func CommonTones(a, b material.Chord) material.PitchClassSet {
	other := b.SoundingPitchClasses()
	var common []material.PitchClass
	for _, pc := range a.SoundingPitchClasses().PitchClasses() {
		if other.Contains(pc) {
			common = append(common, pc)
		}
	}
	return material.NewPitchClassSet(common...)
}

// SecondaryDominantShape reports whether the chord has the shape of a
// secondary dominant in the major key, and returns the degree it points to
// (2 to 6). It is the shape only (docs/architecture/theory-conventions.md, 2):
// naming the chord, and whether the next chord is the one it points to, belong
// to the readings. The chord has the shape when all of these hold:
//
//   - it is a major triad or a dominant seventh: the triad is major and the
//     seventh is none or minor, whatever the tensions;
//   - it has no second, fourth or sixth member: a sus chord, a sixth chord and
//     an add2 chord are not a secondary dominant (D6, D7sus4, Dsus4 and D(add2)
//     in C). This is the owner's call on #103;
//   - a note it is judged by is not in the scale of the key, by pitch class:
//     the root, the third, the fifth, the minor seventh if any, and a bass a
//     minor seventh above the root by spelling, member or not (C/B♭ sounds as
//     C7). The tensions, the second, fourth and sixth members, and any other
//     bass are not looked at, so C(♯11), C(♭6), C/F♯ and C/A♭ are not V/IV in
//     C. This is the owner's call on PR #90;
//   - a perfect fifth below the root, by spelling, is the root of a major or a
//     minor triad of the key other than I. In a major key these are ii, iii,
//     IV, V and vi; vii° is diminished.
//
// It fails for a key other than Ionian (ErrNotMajor), as PositionOf does: the
// degree is counted in the major scale.
func SecondaryDominantShape(c material.Chord, key material.Scale) (target int, ok bool, err error) {
	if key.Mode() != material.Ionian {
		return 0, false, fmt.Errorf("key %s %s: %w", key.Tonic(), key.Mode(), ErrNotMajor)
	}
	if c.Triad() != material.TriadMajor {
		return 0, false, nil
	}
	if s := c.Seventh(); s != material.NoSeventh && s != material.SeventhMinor {
		return 0, false, nil
	}
	if hasSecondFourthOrSixth(c) {
		return 0, false, nil
	}
	var judged []material.PitchClass
	// The chord has no second, fourth or sixth member, so these are all the
	// members the kind of a reading looks at as well (#102), but the tensions.
	for _, n := range notesNumbered(c, 1, 3, 5, 7) {
		judged = append(judged, n.PitchClass())
	}
	if key.Contains(material.NewPitchClassSet(judged...)) {
		return 0, false, nil
	}
	notes := key.Notes()
	// The tonic, degree 1, is left out.
	for degree := 2; degree <= letters; degree++ {
		// An error means no interval of the five qualities, so not a perfect
		// fifth either.
		iv, ivErr := material.IntervalBetween(notes[degree-1], c.Root())
		if ivErr != nil || iv != perfectFifth() {
			continue
		}
		if t := diatonicTriad(notes, degree).Triad(); t != material.TriadMajor && t != material.TriadMinor {
			return 0, false, nil
		}
		return degree, true, nil
	}
	return 0, false, nil
}

// hasSecondFourthOrSixth reports whether the chord has a second, a fourth or a
// sixth member: a sus chord, a sixth chord or an add2 chord.
func hasSecondFourthOrSixth(c material.Chord) bool {
	for _, iv := range c.Intervals() {
		if n := iv.Number(); n == 2 || n == 4 || n == 6 {
			return true
		}
	}
	return false
}

// JudgedNotes returns the spelled notes the kind of a reading compares with
// the scales by: the notes that make the chord, the root and the second to the
// seventh it has (the second and the fourth of sus2 and sus4, the sixth of a
// sixth chord), and a bass a minor seventh above the root by spelling, member
// or not (C/B♭). The tensions (9, 11 and 13) and any other bass are left out
// (C/A♯, whose bass is an augmented sixth). This is the owner's call on #102.
// They come in the order of Chord.Notes, and the bass last.
func JudgedNotes(c material.Chord) []material.NoteName {
	return notesNumbered(c, 1, 2, 3, 4, 5, 6, 7)
}

// notesNumbered returns the members of the interval numbers, and a bass a
// minor seventh above the root, in the order of JudgedNotes. The shape of a
// secondary dominant looks at the root, the third, the fifth and the seventh
// only (the owner's call on PR #90).
func notesNumbered(c material.Chord, numbers ...int) []material.NoteName {
	var out []material.NoteName
	notes := c.Notes()
	for i, iv := range c.Intervals() {
		if slices.Contains(numbers, iv.Number()) {
			out = append(out, notes[i])
		}
	}
	if bass, ok := c.Bass(); ok && bass == minorSeventh() {
		// WithBass checked that the bass is in range.
		n, _ := c.Root().Transpose(bass)
		out = append(out, n)
	}
	return out
}

func perfectFifth() material.Interval {
	iv, err := material.NewInterval(5, material.Perfect)
	if err != nil {
		panic(err) // a perfect fifth is a valid interval
	}
	return iv
}

func minorSeventh() material.Interval {
	iv, err := material.NewInterval(7, material.Minor)
	if err != nil {
		panic(err) // a minor seventh is a valid interval
	}
	return iv
}

// diatonicTriad returns the triad of the scale on the degree: the root, and
// the notes of the scale a third and a fifth above it.
func diatonicTriad(notes []material.NoteName, degree int) material.Chord {
	root := notes[degree-1]
	var intervals []material.Interval
	for _, steps := range []int{2, 4} {
		// The scale is a mode of the naturals moved by one interval, so a
		// third or a fifth between two of its notes is one between naturals,
		// which always has a quality.
		iv, err := material.IntervalBetween(root, notes[(degree-1+steps)%letters])
		if err != nil {
			panic(err)
		}
		intervals = append(intervals, iv)
	}
	c, err := material.NewChord(root, intervals...)
	if err != nil {
		panic(err) // the notes of the scale are in range
	}
	return c
}
