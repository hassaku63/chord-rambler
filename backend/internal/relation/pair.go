package relation

import (
	"fmt"

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
//   - a note it is judged by is not in the scale of the key, by pitch class:
//     the root, the third, the fifth, the minor seventh if any, and a bass a
//     minor seventh above the root by spelling, member or not (C/B♭ sounds as
//     C7). The tensions and any other bass are not looked at, so C(♯11),
//     C/F♯ and C/A♭ are not V/IV in C. This is the owner's call on PR #90;
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
	if key.Contains(judgedPitchClasses(c)) {
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

// judgedPitchClasses returns the pitch classes a major triad or a dominant
// seventh is compared with the key by: the root, the third, the fifth and the
// seventh, and a bass a minor seventh above the root. Other members, such as
// the tensions, and any other bass are left out.
func judgedPitchClasses(c material.Chord) material.PitchClassSet {
	var pcs []material.PitchClass
	notes := c.Notes()
	for i, iv := range c.Intervals() {
		switch iv.Number() {
		case 1, 3, 5, 7:
			pcs = append(pcs, notes[i].PitchClass())
		}
	}
	if bass, ok := c.Bass(); ok && bass == minorSeventh() {
		// WithBass checked that the bass is in range.
		n, _ := c.Root().Transpose(bass)
		pcs = append(pcs, n.PitchClass())
	}
	return material.NewPitchClassSet(pcs...)
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
