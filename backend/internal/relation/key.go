package relation

import (
	"errors"
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrNotMajor reports a key that is not a major scale (Ionian). A position is
// counted from the major scale of the key, and the first scope of this package
// takes major keys only.
var ErrNotMajor = errors.New("key is not a major scale")

// ErrNoPosition reports a chord whose root is two or more half steps away from
// the note of its degree in the major scale, so that no flat, natural or sharp
// degree names it (E♭ in F♯ major, where the seventh degree is E♯).
var ErrNoPosition = errors.New("root is two or more half steps from its degree")

const letters = 7

// tritone is the size of an augmented fourth or a diminished fifth in half
// steps.
const tritone = 6

// PositionOf returns the position of the chord in the major key: the inverse
// of Position.Chord for a position without a target. The degree comes from the
// letters, counted from the tonic up to the root (in C, D♭ is the second
// degree and C♯ the first). The alteration is the difference between the root
// and the note of that degree in the scale: a half step up is Sharp, down is
// Flat. The shape is the chord without its root (material.ShapeOf). No target
// is set; whether a chord leads to another degree is a reading. It fails for a
// key other than Ionian (ErrNotMajor) and for a root two or more half steps
// away from its degree (ErrNoPosition).
func PositionOf(c material.Chord, key material.Scale) (material.Position, error) {
	tonic := key.Tonic()
	if key.Mode() != material.Ionian {
		return material.Position{}, fmt.Errorf("key %s %s: %w", tonic, key.Mode(), ErrNotMajor)
	}
	root := c.Root()
	steps := (int(root.Letter()) - int(tonic.Letter()) + letters) % letters
	// Both notes are on the same letter, so the difference of their
	// alterations is the difference in half steps.
	alter := root.Alter() - key.Notes()[steps].Alter()
	if alter < int(material.Flat) || alter > int(material.Sharp) {
		return material.Position{}, fmt.Errorf("root %s in %s major: %w", root, tonic, ErrNoPosition)
	}
	return material.NewPosition(steps+1, material.Alteration(alter), material.ShapeOf(c))
}

// InScale reports whether the pitch classes that sound (the members and the
// bass, Chord.SoundingPitchClasses) are all in the scale of the key.
func InScale(c material.Chord, key material.Scale) bool {
	return key.Contains(c.SoundingPitchClasses())
}

// ParallelModes returns the modes on the tonic of the key, of the seven church
// modes, whose scale contains the pitch classes that sound. They come in the
// order of material.Mode, Ionian first. Only the tonic of the key is used, and
// the scales are compared by pitch class, so a mode whose spelling on the
// tonic would need more than a double sharp or a double flat is still
// compared.
func ParallelModes(c material.Chord, key material.Scale) []material.Mode {
	sounding := c.SoundingPitchClasses()
	shift := key.Tonic().PitchClass().Int()
	var modes []material.Mode
	for m := material.Ionian; m <= material.Locrian; m++ {
		// Every mode on C is spelled within a single sharp or flat.
		onC, _ := material.NewScale(material.NoteName{}, m)
		if sounding.IsSubsetOf(onC.PitchClasses().Transpose(shift)) {
			modes = append(modes, m)
		}
	}
	return modes
}

// HasLeadingTone reports whether the pitch classes that sound contain the
// leading tone of the key, the note a minor second below the tonic.
func HasLeadingTone(c material.Chord, key material.Scale) bool {
	leading := material.NewPitchClassSet(key.Tonic().PitchClass()).Transpose(-1)
	return leading.IsSubsetOf(c.SoundingPitchClasses())
}

// HasTritone reports whether two of the pitch classes that sound are six half
// steps apart (an augmented fourth or a diminished fifth, by pitch class).
func HasTritone(c material.Chord) bool {
	sounding := c.SoundingPitchClasses()
	for _, pc := range sounding.PitchClasses() {
		if material.NewPitchClassSet(pc).Transpose(tritone).IsSubsetOf(sounding) {
			return true
		}
	}
	return false
}
