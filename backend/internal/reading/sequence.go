package reading

import (
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// Resolution is how a dominant goes on to the next chord
// (docs/architecture/theory-conventions.md, 2). The zero value, NoResolution,
// says nothing.
type Resolution int

// The resolutions.
const (
	// NoResolution is said of every chord that is neither V nor a secondary
	// dominant, and of one that no chord follows.
	NoResolution Resolution = iota
	// ResolutionExpected is a dominant followed by the chord it points to.
	ResolutionExpected
	// ResolutionIrregular is a dominant followed by another chord.
	ResolutionIrregular
)

// String returns the name of the constant, for tests and logs.
func (r Resolution) String() string {
	switch r {
	case NoResolution:
		return "NoResolution"
	case ResolutionExpected:
		return "ResolutionExpected"
	case ResolutionIrregular:
		return "ResolutionIrregular"
	default:
		return fmt.Sprintf("Resolution(%d)", int(r))
	}
}

// Cadence is the name given to the end of a phrase
// (docs/architecture/theory-conventions.md, 4). The zero value, NoCadence,
// gives none.
type Cadence int

// The cadences.
const (
	NoCadence Cadence = iota
	// AuthenticCadence is V to I.
	AuthenticCadence
	// HalfCadence is a phrase ending on V.
	HalfCadence
	// PlagalCadence is IV to I.
	PlagalCadence
	// DeceptiveCadence is V to a chord other than I.
	DeceptiveCadence
)

// String returns the name of the constant, for tests and logs.
func (c Cadence) String() string {
	switch c {
	case NoCadence:
		return "NoCadence"
	case AuthenticCadence:
		return "AuthenticCadence"
	case HalfCadence:
		return "HalfCadence"
	case PlagalCadence:
		return "PlagalCadence"
	case DeceptiveCadence:
		return "DeceptiveCadence"
	default:
		return fmt.Sprintf("Cadence(%d)", int(c))
	}
}

// SlotReading is the reading of one slot of a progression.
type SlotReading struct {
	// HasChord reports whether the slot holds a chord. A slot with no chord,
	// or none chosen yet, has no reading.
	HasChord bool
	// Chord is the reading of the chord of the slot (ReadChord), set when
	// HasChord is true.
	Chord ChordReading
	// Resolution is how the chord goes on to the chord of the next slot.
	Resolution Resolution
	// Cadence is set on the last slot of a phrase only.
	Cadence Cadence
}

// ReadProgression reads every slot of the progression in its key, in order:
// the element at i is the slot at i, as relation.Places numbers them.
//
// A chord is taken as a degree of the key when it is read as diatonic
// (ReadChord) and its root is on that degree (the owner's call on #103): in
// C, G, G7, G/B and G7(♭9) are on the fifth degree, Gm (borrowed) and D7 (a
// secondary dominant) are not on any. V is a chord on the fifth degree whose
// triad is major and whose seventh is none or minor, whatever its bass and
// tensions (G/B, G7/F, G7(9); not Gsus4).
//
// Resolution is said of V and of a secondary dominant, when the next slot
// holds a chord: expected when that chord is on the degree pointed to (1 for
// V, the target for a secondary dominant), irregular otherwise. The next slot
// may be in the next phrase.
//
// A cadence is named on the last slot of a phrase only:
//
//   - half, when the slot holds V, whatever comes before it;
//   - otherwise only when the slot before it is in the same phrase and both
//     hold chords: authentic when V goes to I, deceptive when V goes to any
//     other chord, plagal when IV goes to I.
//
// It fails for a key other than Ionian (relation.ErrNotMajor).
func ReadProgression(p progression.Progression) ([]SlotReading, error) {
	key := p.Key()
	if key.Mode() != material.Ionian {
		return nil, fmt.Errorf("read progression in %s %s: %w", key.Tonic(), key.Mode(), relation.ErrNotMajor)
	}
	slots := p.Slots()
	read := make([]slotChord, len(slots))
	for i, s := range slots {
		r, err := readSlot(s, key)
		if err != nil {
			return nil, err
		}
		read[i] = r
	}
	places := relation.Places(p)
	out := make([]SlotReading, len(slots))
	for i, r := range read {
		out[i] = SlotReading{HasChord: r.ok, Chord: r.reading}
		if i+1 < len(read) {
			out[i].Resolution = resolution(r, read[i+1])
		}
		if places[i].LastInPhrase {
			var prev slotChord
			if places[i].IndexInPhrase > 1 {
				prev = read[i-1]
			}
			out[i].Cadence = cadence(prev, r)
		}
	}
	return out, nil
}

// slotChord is what the rules look at in a slot: whether it holds a chord, the
// reading, the degree of a chord read as diatonic (0 otherwise), and whether
// the chord is V.
type slotChord struct {
	ok      bool
	reading ChordReading
	degree  int
	v       bool
}

func readSlot(s progression.Slot, key material.Scale) (slotChord, error) {
	c, ok := s.Content().Chord()
	if !ok {
		return slotChord{}, nil
	}
	r, err := ReadChord(c, key)
	if err != nil {
		return slotChord{}, fmt.Errorf("read progression: %w", err)
	}
	out := slotChord{ok: true, reading: r}
	if r.Main.Kind == Diatonic {
		pos, err := relation.PositionOf(c, key)
		if err != nil {
			// Not reached: the key is Ionian and the root of a diatonic chord
			// is a note of its scale.
			panic(err)
		}
		out.degree = pos.Degree()
	}
	seventh := c.Seventh()
	out.v = out.degree == 5 && c.Triad() == material.TriadMajor &&
		(seventh == material.NoSeventh || seventh == material.SeventhMinor)
	return out, nil
}

// resolution returns how the chord goes on to the next.
func resolution(cur, next slotChord) Resolution {
	var want int
	switch {
	case !cur.ok || !next.ok:
		return NoResolution
	case cur.reading.Main.Kind == SecondaryDominant:
		want = cur.reading.Main.Target
	case cur.v:
		want = 1
	default:
		return NoResolution
	}
	if next.degree == want {
		return ResolutionExpected
	}
	return ResolutionIrregular
}

// cadence returns the cadence that ends a phrase on last, after prev, the
// slot before it in the phrase (the zero value when there is none).
func cadence(prev, last slotChord) Cadence {
	switch {
	case !last.ok:
		return NoCadence
	case last.v:
		return HalfCadence
	case !prev.ok:
		return NoCadence
	case prev.v && last.degree == 1:
		return AuthenticCadence
	case prev.v:
		return DeceptiveCadence
	case prev.degree == 4 && last.degree == 1:
		return PlagalCadence
	default:
		return NoCadence
	}
}
