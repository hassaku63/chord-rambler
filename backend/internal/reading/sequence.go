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
	// dominant, of one that no chord follows, of one followed by the same
	// dominant, and of one followed by a chord that is not a triad with a
	// third or a seventh chord (a sus or a sixth chord).
	NoResolution Resolution = iota
	// ResolutionExpected is a dominant followed by the chord it points to.
	ResolutionExpected
	// ResolutionIrregular is a dominant followed by a triad with a third or a
	// seventh chord that is not the one it points to.
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
	// Cadence is set on the last slot of a full phrase only.
	Cadence Cadence
}

// ReadProgression reads every slot of the progression in its key, in order:
// the element at i is the slot at i, as relation.Places numbers them.
//
// A chord is on a degree of the key when it is read as diatonic (ReadChord),
// its root is on that degree, and it is a triad with a third or a seventh
// chord, whatever its bass and tensions: no second, fourth or sixth member
// (the owner's calls on #103). In C, G, G7, G/B and G7(♭9) are on the fifth
// degree; Gsus4, G6, Gm (borrowed) and D7 (a secondary dominant) are not. V is
// a chord on the fifth degree whose triad is major and whose seventh is none
// or minor; I and IV are chords on the first and the fourth degrees.
//
// Resolution is said of V and of a secondary dominant when the next slot
// holds a triad with a third or a seventh chord: expected when that chord is
// on the degree pointed to (1 for V, the target for a secondary dominant),
// irregular otherwise. It is not said when the same dominant goes on (V to V,
// or a secondary dominant to one with the same target): the last of them
// says it. The next slot may be in the next phrase.
//
// A cadence is named only on the last slot of a phrase that is full (as long
// as progression.Progression.PhraseBeats), so a phrase still being written at
// the end gets none:
//
//   - half, when the slot holds V, whatever comes before it;
//   - otherwise only when the slot holds a triad with a third or a seventh
//     chord, and the slot before it is in the same phrase and holds a chord:
//     authentic when V goes to I, deceptive when V goes to any other chord,
//     plagal when IV goes to I.
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
	end := 0 // beats from the start of the progression to the end of the slot
	for i, r := range read {
		end += slots[i].Beats()
		out[i] = SlotReading{HasChord: r.ok, Chord: r.reading}
		if i+1 < len(read) {
			out[i].Resolution = resolution(r, read[i+1])
		}
		// No slot crosses a phrase, so a phrase is full when its last slot
		// ends on a multiple of the length of a phrase.
		if places[i].LastInPhrase && end%p.PhraseBeats() == 0 {
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
// reading, the degree of a chord read as diatonic (0 otherwise), whether the
// chord is a triad with a third or a seventh chord (tertian), and whether it
// is V.
type slotChord struct {
	ok      bool
	reading ChordReading
	degree  int
	tertian bool
	v       bool
}

// on reports whether the chord is on the degree: read as diatonic there, and
// tertian.
func (s slotChord) on(degree int) bool { return s.tertian && s.degree == degree }

// target returns the degree a dominant points to (1 for V), or 0.
func (s slotChord) target() int {
	switch {
	case s.reading.Main.Kind == SecondaryDominant:
		return s.reading.Main.Target
	case s.v:
		return 1
	default:
		return 0
	}
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
	out := slotChord{ok: true, reading: r, tertian: tertian(c)}
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
	out.v = out.on(5) && c.Triad() == material.TriadMajor &&
		(seventh == material.NoSeventh || seventh == material.SeventhMinor)
	return out, nil
}

// tertian reports whether the chord is a triad with a third (major, minor,
// diminished or augmented), with or without a seventh and tensions, and with
// no second, fourth or sixth member. The bass is not looked at.
func tertian(c material.Chord) bool {
	switch c.Triad() {
	case material.TriadMajor, material.TriadMinor, material.TriadDiminished, material.TriadAugmented:
	default:
		return false
	}
	for _, iv := range c.Intervals() {
		if n := iv.Number(); n == 2 || n == 4 || n == 6 {
			return false
		}
	}
	return true
}

// resolution returns how the chord goes on to the next.
func resolution(cur, next slotChord) Resolution {
	want := cur.target()
	switch {
	case !next.ok || want == 0 || !next.tertian:
		return NoResolution
	case next.target() == want && next.reading.Main.Kind == cur.reading.Main.Kind:
		// The same dominant goes on: V to V, or V/x to V/x.
		return NoResolution
	case next.on(want):
		return ResolutionExpected
	default:
		return ResolutionIrregular
	}
}

// cadence returns the cadence that ends a full phrase on last, after prev, the
// slot before it in the phrase (the zero value when there is none).
func cadence(prev, last slotChord) Cadence {
	switch {
	case !last.ok:
		return NoCadence
	case last.v:
		return HalfCadence
	case !last.tertian || !prev.ok:
		return NoCadence
	case prev.v && last.on(1):
		return AuthenticCadence
	case prev.v:
		return DeceptiveCadence
	case prev.on(4) && last.on(1):
		return PlagalCadence
	default:
		return NoCadence
	}
}
