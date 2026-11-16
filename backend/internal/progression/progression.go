package progression

import (
	"errors"
	"fmt"
	"slices"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
)

// ErrIndex reports a slot index outside the progression.
var ErrIndex = errors.New("slot index is outside the progression")

// ErrRangeLength reports a range replaced by slots whose total length differs
// from that of the range.
var ErrRangeLength = errors.New("slots do not add up to the length of the range")

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
	length := p.PhraseBeats()
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

// PhraseBeats returns the length of a full phrase in beats: four bars of the
// meter. Every phrase but the last is full.
func (p Progression) PhraseBeats() int { return phraseBars * p.meter.BeatsPerBar() }

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

// Fill returns the progression with the content of the slot at the index (from
// 0) replaced. The length and the role of the slot are kept, so the phrases do
// not change. It fails for an index outside 0 to Len()-1 (ErrIndex).
func (p Progression) Fill(i int, c Content) (Progression, error) {
	if i < 0 || i >= len(p.slots) {
		return Progression{}, fmt.Errorf("fill slot %d of %d: %w", i, len(p.slots), ErrIndex)
	}
	slots := slices.Clone(p.slots)
	slots[i] = slots[i].WithContent(c)
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

// ReplaceRange returns the progression with the slots from index i up to but
// not including index j replaced, in one step, by the slots given. The slots
// given must add up to the same number of beats as the range, so every slot
// outside the range starts where it did. This splits a slot, joins slots and
// shares beats out between them in any phrase. It fails for a range outside
// the progression or an empty one, that is unless 0 <= i < j <= Len()
// (ErrIndex), for slots of another total length (ErrRangeLength), and when a
// slot given would cross the boundary of a phrase (ErrCrossesPhrase).
func (p Progression) ReplaceRange(i, j int, slots ...Slot) (Progression, error) {
	if i < 0 || j > len(p.slots) || i >= j {
		return Progression{}, fmt.Errorf("replace slots %d to %d of %d: %w", i, j, len(p.slots), ErrIndex)
	}
	if have, give := totalBeats(p.slots[i:j]), totalBeats(slots); have != give {
		return Progression{}, fmt.Errorf("replace %d beats with %d: %w", have, give, ErrRangeLength)
	}
	return p.with(slices.Concat(p.slots[:i], slots, p.slots[j:]))
}

func totalBeats(slots []Slot) int {
	sum := 0
	for _, s := range slots {
		sum += s.Beats()
	}
	return sum
}

// with returns the progression with the slots, which the caller does not
// share, or fails when a slot would cross the boundary of a phrase. Every
// operation that changes the slots goes through it.
func (p Progression) with(slots []Slot) (Progression, error) {
	length := p.PhraseBeats()
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

// Transpose returns the progression with the tonic of the key and the root of
// every chord moved up by the interval, as material's Transpose moves a note
// name, so the spelling is kept. Each chord keeps its intervals and its bass.
// Slots with no chord or none chosen stay as they are, and every length and
// role stays. It fails, changing nothing, when the tonic or a chord would need
// more than a double sharp or a double flat (material.ErrAlterOutOfRange).
func (p Progression) Transpose(iv material.Interval) (Progression, error) {
	return p.move(func(n material.NoteName) (material.NoteName, error) { return n.Transpose(iv) })
}

// TransposeDown is Transpose with the interval downward. Transposing up and
// then down by the same interval gives the progression back.
func (p Progression) TransposeDown(iv material.Interval) (Progression, error) {
	return p.move(func(n material.NoteName) (material.NoteName, error) { return n.TransposeDown(iv) })
}

func (p Progression) move(by func(material.NoteName) (material.NoteName, error)) (Progression, error) {
	tonic, err := by(p.key.Tonic())
	if err != nil {
		return Progression{}, fmt.Errorf("tonic: %w", err)
	}
	key, err := material.NewScale(tonic, p.key.Mode())
	if err != nil {
		return Progression{}, fmt.Errorf("key: %w", err)
	}
	slots := slices.Clone(p.slots)
	for i, s := range slots {
		c, ok := s.content.Chord()
		if !ok {
			continue
		}
		root, err := by(c.Root())
		if err != nil {
			return Progression{}, fmt.Errorf("slot %d: %w", i, err)
		}
		moved, err := material.ShapeOf(c).On(root)
		if err != nil {
			return Progression{}, fmt.Errorf("slot %d: %w", i, err)
		}
		slots[i].content = ChordOf(moved)
	}
	p.key, p.slots = key, slots
	return p, nil
}
