package progression

import (
	"errors"
	"fmt"
	"slices"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrIndex reports a slot index outside the progression.
var ErrIndex = errors.New("slot index is outside the progression")

// ErrCrossesPhrase reports an operation whose result would have a slot that
// crosses the boundary of a phrase. A slot may cross the boundary of a bar,
// but not of a phrase (#91).
var ErrCrossesPhrase = errors.New("slot crosses the boundary of a phrase")

// phraseBars is the length of a phrase in bars. Phrases are cut every four
// bars from the start; the last one may be shorter.
const phraseBars = 4

// Progression is a key, a meter and the slots, nested as one section of
// phrases of slots. Phrases are cut every four bars from the start, and the
// last phrase may be shorter. No slot crosses the boundary of a phrase.
// Operations return a new progression and leave the one they are called on as
// it was. When an operation fails, do not use the progression it returns.
// The zero value is C major (C Ionian), 4/4, with no slots. Build
// other values with New and the operations.
type Progression struct {
	key   material.Scale
	meter Meter
	slots []Slot
}

// New returns the progression of the key and the meter with no slots.
func New(key material.Scale, m Meter) Progression {
	return Progression{key: key, meter: m}
}

// Key returns the key.
func (p Progression) Key() material.Scale { return p.key }

// Meter returns the meter.
func (p Progression) Meter() Meter { return p.meter }

// Len returns the number of slots.
func (p Progression) Len() int { return len(p.slots) }

// Slots returns the slots in order, through every phrase.
func (p Progression) Slots() []Slot { return slices.Clone(p.slots) }

// Section is a large part of a song (a verse, a chorus) made of phrases.
type Section struct {
	phrases []Phrase
}

// Phrases returns the phrases in order.
func (s Section) Phrases() []Phrase { return slices.Clone(s.phrases) }

// Phrase is one cut of the progression, four bars by default.
type Phrase struct {
	slots []Slot
}

// Slots returns the slots of the phrase in order.
func (ph Phrase) Slots() []Slot { return slices.Clone(ph.slots) }

// Sections returns the sections. The first scope has one section, which holds
// every phrase; a progression with no slots has one section with no phrases.
func (p Progression) Sections() []Section {
	var phrases []Phrase
	length := p.phraseBeats()
	start := 0
	for _, s := range p.slots {
		if start%length == 0 {
			phrases = append(phrases, Phrase{})
		}
		last := &phrases[len(phrases)-1]
		last.slots = append(last.slots, s)
		start += s.Beats()
	}
	return []Section{{phrases: phrases}}
}

func (p Progression) phraseBeats() int { return phraseBars * p.meter.BeatsPerBar() }

// Append returns the progression with the slot added at the end.
func (p Progression) Append(s Slot) (Progression, error) {
	return p.Insert(len(p.slots), s)
}

// Replace returns the progression with the slot at the index (from 0) replaced.
// It fails for an index outside 0 to Len()-1 (ErrIndex).
func (p Progression) Replace(i int, s Slot) (Progression, error) {
	if i < 0 || i >= len(p.slots) {
		return Progression{}, fmt.Errorf("replace slot %d of %d: %w", i, len(p.slots), ErrIndex)
	}
	slots := slices.Clone(p.slots)
	slots[i] = s
	return p.with(slots)
}

// Insert returns the progression with the slot inserted before the index
// (from 0). An index of Len() adds the slot at the end, as Append does. It
// fails for an index outside 0 to Len() (ErrIndex).
func (p Progression) Insert(i int, s Slot) (Progression, error) {
	if i < 0 || i > len(p.slots) {
		return Progression{}, fmt.Errorf("insert before slot %d of %d: %w", i, len(p.slots), ErrIndex)
	}
	return p.with(slices.Insert(slices.Clone(p.slots), i, s))
}

// Delete returns the progression without the slot at the index (from 0). It
// fails for an index outside 0 to Len()-1 (ErrIndex).
func (p Progression) Delete(i int) (Progression, error) {
	if i < 0 || i >= len(p.slots) {
		return Progression{}, fmt.Errorf("delete slot %d of %d: %w", i, len(p.slots), ErrIndex)
	}
	return p.with(slices.Delete(slices.Clone(p.slots), i, i+1))
}

// with returns the progression with the slots, which the caller does not
// share, or fails when a slot would cross the boundary of a phrase. Every
// operation that changes the slots goes through it.
func (p Progression) with(slots []Slot) (Progression, error) {
	length := p.phraseBeats()
	start := 0
	for i, s := range slots {
		end := start + s.Beats()
		if start/length != (end-1)/length {
			return Progression{}, fmt.Errorf("slot %d (beats %d to %d, a phrase every %d beats): %w",
				i, start, end, length, ErrCrossesPhrase)
		}
		start = end
	}
	p.slots = slots
	return p, nil
}
