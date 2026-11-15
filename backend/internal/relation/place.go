package relation

import "github.com/hassaku63/chord-rambler/backend/internal/progression"

// Place is where a slot sits in time: its bar and beat, its phrase, and
// whether it ends the phrase or the progression. Every number counts from 1.
type Place struct {
	// Bar and Beat are where the slot starts. A slot that crosses a bar line
	// is placed by the bar and the beat it starts on.
	Bar, Beat int
	// Phrase is the phrase the slot belongs to. No slot crosses the boundary
	// of a phrase, so there is exactly one.
	Phrase int
	// IndexInPhrase is the position of the slot among the slots of its phrase.
	IndexInPhrase int
	// LastInPhrase and LastInProgression report whether no slot follows in the
	// phrase and in the progression.
	LastInPhrase, LastInProgression bool
}

// Places returns the place of every slot of the progression, in order: the
// element at i is the slot at i (from 0, as the progression numbers them). A
// progression with no slots has no places.
//
// Only the facts are computed. Whether the end of a phrase is a cadence, and
// what role a slot plays, belong to the readings.
func Places(p progression.Progression) []Place {
	perBar := p.Meter().BeatsPerBar()
	var out []Place
	start := 0 // beats before the slot, from the start of the progression
	for i, ph := range phrases(p) {
		slots := ph.Slots()
		for j, s := range slots {
			out = append(out, Place{
				Bar:           start/perBar + 1,
				Beat:          start%perBar + 1,
				Phrase:        i + 1,
				IndexInPhrase: j + 1,
				LastInPhrase:  j == len(slots)-1,
			})
			start += s.Beats()
		}
	}
	if n := len(out); n > 0 {
		out[n-1].LastInProgression = true
	}
	return out
}

// phrases returns the phrases of the progression through its sections, in
// order.
func phrases(p progression.Progression) []progression.Phrase {
	var out []progression.Phrase
	for _, sec := range p.Sections() {
		out = append(out, sec.Phrases()...)
	}
	return out
}
