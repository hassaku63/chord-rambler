package reading_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// slot is a slot of a test progression: its content and its length in beats.
type slot struct {
	content progression.Content
	beats   int
}

// bar is a slot of one bar of 4/4 holding the chord.
func bar(c material.Chord) slot { return slot{progression.ChordOf(c), 4} }

// Slots of one bar that hold no chord, or none chosen yet.
var (
	noChord  = slot{progression.NoChord(), 4}
	unchosen = slot{progression.Unchosen(), 4}
)

// build appends the slots to an empty progression in the key, in 4/4.
func build(key material.Scale, slots ...slot) progression.Progression {
	p := progression.New(key, progression.Meter{})
	for _, s := range slots {
		ps, err := progression.NewSlot(s.content, s.beats)
		if err != nil {
			panic(err)
		}
		if p, err = p.Append(ps); err != nil {
			panic(err)
		}
	}
	return p
}

// Chords of C major for the tables, as chord symbols.
var (
	c       = chord(cNat, maj, "")
	cMaj7   = chord(cNat, maj7, "")
	cOverE  = chord(cNat, maj, "M3")
	cm      = chord(cNat, minTriad, "")
	c7      = chord(cNat, dom7, "")
	dm      = chord(dNat, minTriad, "")
	d7      = chord(dNat, dom7, "")
	e       = chord(eNat, maj, "")
	em      = chord(eNat, minTriad, "")
	f       = chord(fNat, maj, "")
	g       = chord(gNat, maj, "")
	g7      = chord(gNat, dom7, "")
	gOverB  = chord(gNat, maj, "M3")
	g7b9    = chord(gNat, "M3 P5 m7 m9", "")
	g7OverF = chord(gNat, dom7, "m7")
	g7n9    = chord(gNat, "M3 P5 m7 M9", "")
	gsus4   = chord(gNat, "P4 P5", "")
	gm      = chord(gNat, minTriad, "")
	am      = chord(aNat, minTriad, "")
	a7      = chord(aNat, dom7, "")
	b7      = chord(bNat, dom7, "")
)

func readProgression(t *testing.T, p progression.Progression) []reading.SlotReading {
	t.Helper()
	got, err := reading.ReadProgression(p)
	if err != nil {
		t.Fatalf("ReadProgression: %v", err)
	}
	if len(got) != p.Len() {
		t.Fatalf("ReadProgression: %d readings for %d slots", len(got), p.Len())
	}
	return got
}

// How a dominant goes on, read on the first of two slots in C major
// (docs/architecture/theory-conventions.md, 2; Hutchinson 17.4,
// https://musictheory.pugetsound.edu/mt21c/AnalyzingSecondaryDominants.html,
// and 17.6,
// https://musictheory.pugetsound.edu/mt21c/IrregularResoltionsOfSecondaryDominants.html).
// The chord is expected when the next is read as diatonic and is on the degree
// pointed to (the owner's call on #103): G7, G/B and G7(♭9) after D7, but not
// Gm (borrowed) or D7 after A7 (a secondary dominant). V is the major triad or
// the dominant seventh on the fifth degree, whatever its bass and tensions
// (#103): G/B, G7/F and G7(9) are V, Gsus4 is not. Nothing is said when no
// chord follows.
func TestResolution(t *testing.T) {
	tests := []struct {
		name  string
		slots []slot
		want  reading.Resolution
	}{
		// The cases of #103.
		{"D7 → G", []slot{bar(d7), bar(g)}, reading.ResolutionExpected},
		{"D7 → Em", []slot{bar(d7), bar(em)}, reading.ResolutionIrregular},
		{"G7 → C", []slot{bar(g7), bar(c)}, reading.ResolutionExpected},
		{"G7 → Am", []slot{bar(g7), bar(am)}, reading.ResolutionIrregular},
		// The owner's call on #103.
		{"D7 → G7", []slot{bar(d7), bar(g7)}, reading.ResolutionExpected},
		{"D7 → G/B", []slot{bar(d7), bar(gOverB)}, reading.ResolutionExpected},
		{"D7 → G7(♭9)", []slot{bar(d7), bar(g7b9)}, reading.ResolutionExpected},
		{"D7 → Gm", []slot{bar(d7), bar(gm)}, reading.ResolutionIrregular},
		{"A7 → D7", []slot{bar(a7), bar(d7)}, reading.ResolutionIrregular},
		{"G/B → C", []slot{bar(gOverB), bar(c)}, reading.ResolutionExpected},
		{"G7/F → C", []slot{bar(g7OverF), bar(c)}, reading.ResolutionExpected},
		{"G7(9) → C", []slot{bar(g7n9), bar(c)}, reading.ResolutionExpected},
		{"Gsus4 → C: not V", []slot{bar(gsus4), bar(c)}, reading.NoResolution},
		// What follows from the calls.
		{"D7 → Gsus4: diatonic on the fifth degree", []slot{bar(d7), bar(gsus4)}, reading.ResolutionExpected},
		{"A7 → Dm", []slot{bar(a7), bar(dm)}, reading.ResolutionExpected},
		{"B7 → Em", []slot{bar(b7), bar(em)}, reading.ResolutionExpected},
		{"C7 → F", []slot{bar(c7), bar(f)}, reading.ResolutionExpected},
		{"E → Am", []slot{bar(e), bar(am)}, reading.ResolutionExpected},
		{"G7 → Cm: borrowed", []slot{bar(g7), bar(cm)}, reading.ResolutionIrregular},
		{"G7 → C7: a secondary dominant", []slot{bar(g7), bar(c7)}, reading.ResolutionIrregular},
		{"C → F: not a dominant", []slot{bar(c), bar(f)}, reading.NoResolution},
		{"Am → Dm: not a dominant", []slot{bar(am), bar(dm)}, reading.NoResolution},
		{"D7 last", []slot{bar(d7)}, reading.NoResolution},
		{"D7 → no chord", []slot{bar(d7), noChord}, reading.NoResolution},
		{"D7 → none chosen", []slot{bar(d7), unchosen}, reading.NoResolution},
		{"G7 → no chord", []slot{bar(g7), noChord}, reading.NoResolution},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readProgression(t, build(major(cNat), tt.slots...))
			if got[0].Resolution != tt.want {
				t.Errorf("Resolution = %s; want %s", got[0].Resolution, tt.want)
			}
		})
	}
}

// The cadences of every slot of progressions in C major. A cadence is named
// on the last slot of a phrase only (docs/architecture/theory-conventions.md,
// 4; Hutchinson 13.1, https://musictheory.pugetsound.edu/mt21c/cadences.html):
// authentic V → I, half on V, plagal IV → I, deceptive V → other than I. V,
// I and IV are read as in TestResolution (the owner's calls on #103): G7 → Cm
// is deceptive. A phrase ending on V is a half cadence whatever comes before
// it; the other cadences need a chord in the slot before, in the same phrase
// (decided in #103, not by the owner).
func TestCadence(t *testing.T) {
	const (
		none      = reading.NoCadence
		authentic = reading.AuthenticCadence
		half      = reading.HalfCadence
		plagal    = reading.PlagalCadence
		deceptive = reading.DeceptiveCadence
	)
	tests := []struct {
		name  string
		slots []slot
		want  []reading.Cadence
	}{
		// The cases of #103: four bars, one phrase.
		{"authentic: G7 → C", []slot{bar(c), bar(f), bar(g7), bar(c)}, []reading.Cadence{none, none, none, authentic}},
		{"half: ends on G", []slot{bar(c), bar(am), bar(f), bar(g)}, []reading.Cadence{none, none, none, half}},
		{"plagal: F → C", []slot{bar(c), bar(am), bar(f), bar(c)}, []reading.Cadence{none, none, none, plagal}},
		{"deceptive: G → Am", []slot{bar(c), bar(f), bar(g), bar(am)}, []reading.Cadence{none, none, none, deceptive}},
		{"G7 → C in the middle of the phrase", []slot{bar(c), bar(g7), bar(c), bar(am)}, []reading.Cadence{none, none, none, none}},
		// The owner's call on #103.
		{"G7 → Cmaj7", []slot{bar(f), bar(g7), bar(cMaj7)}, []reading.Cadence{none, none, authentic}},
		{"G7 → C/E", []slot{bar(f), bar(g7), bar(cOverE)}, []reading.Cadence{none, none, authentic}},
		{"G7 → Cm", []slot{bar(f), bar(g7), bar(cm)}, []reading.Cadence{none, none, deceptive}},
		{"G/B → C", []slot{bar(f), bar(gOverB), bar(c)}, []reading.Cadence{none, none, authentic}},
		{"G7/F → C", []slot{bar(f), bar(g7OverF), bar(c)}, []reading.Cadence{none, none, authentic}},
		{"G7(9) → C", []slot{bar(f), bar(g7n9), bar(c)}, []reading.Cadence{none, none, authentic}},
		{"Gsus4 → C: not V", []slot{bar(f), bar(gsus4), bar(c)}, []reading.Cadence{none, none, none}},
		{"ends on Gsus4: not V", []slot{bar(c), bar(f), bar(gsus4)}, []reading.Cadence{none, none, none}},
		// What follows from the calls.
		{"G7 → C7: a secondary dominant is not I", []slot{bar(f), bar(g7), bar(c7)}, []reading.Cadence{none, none, deceptive}},
		{"D7 → G: ends on V", []slot{bar(c), bar(d7), bar(g)}, []reading.Cadence{none, none, half}},
		{"Dm → C: not IV", []slot{bar(c), bar(dm), bar(c)}, []reading.Cadence{none, none, none}},
		{"one slot, C", []slot{bar(c)}, []reading.Cadence{none}},
		{"one slot, G", []slot{bar(g)}, []reading.Cadence{half}},
		{
			"each phrase ends on its own",
			[]slot{bar(c), bar(f), bar(g7), bar(c), bar(am), bar(dm), bar(g7), bar(am)},
			[]reading.Cadence{none, none, none, authentic, none, none, none, deceptive},
		},
		// Slots with no chord (decided in #103, not by the owner).
		{"F → no chord → C", []slot{bar(c), bar(f), noChord, bar(c)}, []reading.Cadence{none, none, none, none}},
		{"G7 → none chosen → C", []slot{bar(c), bar(g7), unchosen, bar(c)}, []reading.Cadence{none, none, none, none}},
		{"G → no chord at the end", []slot{bar(c), bar(f), bar(g), noChord}, []reading.Cadence{none, none, none, none}},
		{"G → none chosen at the end", []slot{bar(c), bar(f), bar(g), unchosen}, []reading.Cadence{none, none, none, none}},
		{"no chord → G at the end", []slot{bar(c), bar(f), noChord, bar(g)}, []reading.Cadence{none, none, none, half}},
		// A phrase of one slot has no slot before it in the phrase (decided in
		// #103, not by the owner).
		{
			"G7 | C of four bars",
			[]slot{bar(c), bar(f), bar(dm), bar(g7), {progression.ChordOf(c), 16}},
			[]reading.Cadence{none, none, none, half, none},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readProgression(t, build(major(cNat), tt.slots...))
			var cadences []reading.Cadence
			for _, r := range got {
				cadences = append(cadences, r.Cadence)
			}
			if !slices.Equal(cadences, tt.want) {
				t.Errorf("cadences = %v; want %v", cadences, tt.want)
			}
		})
	}
}

// Each slot holding a chord carries the reading of ReadChord in the key of the
// progression; the others carry none. A progression with no slots has no
// readings.
func TestReadProgressionCarriesTheChordReadings(t *testing.T) {
	key := major(note(material.F, sharp))
	chords := []material.Chord{
		chord(note(material.C, sharp), dom7, ""), // V7
		chord(eFlat, maj, ""),                    // no position in F♯
		chord(note(material.G, sharp), dom7, ""), // V7/V
	}
	got := readProgression(t, build(key, bar(chords[0]), noChord, bar(chords[1]), unchosen, bar(chords[2])))
	for i, ci := range []int{0, -1, 1, -1, 2} {
		if ci < 0 {
			if got[i].HasChord || !isZero(got[i].Chord) {
				t.Errorf("slot %d: HasChord, Chord = %v, %+v; want false and nothing", i, got[i].HasChord, got[i].Chord)
			}
			continue
		}
		want, err := reading.ReadChord(chords[ci], key)
		if err != nil {
			t.Fatal(err)
		}
		if !got[i].HasChord || got[i].Chord.Main.Kind != want.Main.Kind ||
			got[i].Chord.Main.Function != want.Main.Function || got[i].Chord.Main.Target != want.Main.Target ||
			got[i].Chord.Main.Key != key {
			t.Errorf("slot %d: HasChord, Chord = %v, %+v; want true and %+v", i, got[i].HasChord, got[i].Chord, want)
		}
	}
	if empty := readProgression(t, build(major(cNat))); len(empty) != 0 {
		t.Errorf("ReadProgression of no slots = %v; want none", empty)
	}
}

func isZero(r reading.ChordReading) bool {
	return r.Main.Kind == reading.FactsOnly && r.Main.Function == reading.NoFunction &&
		r.Main.Target == 0 && r.Main.Modes == nil && r.Others == nil && r.Main.Key == material.Scale{}
}

// A key other than Ionian is refused, even with no chord to read.
func TestReadProgressionOtherThanMajor(t *testing.T) {
	for m := material.Dorian; m <= material.Locrian; m++ {
		key, err := material.NewScale(cNat, m)
		if err != nil {
			t.Fatal(err)
		}
		for _, slots := range [][]slot{nil, {bar(g7), bar(c)}} {
			if _, err := reading.ReadProgression(build(key, slots...)); !errors.Is(err, relation.ErrNotMajor) {
				t.Errorf("ReadProgression of %d slots in C %s: err = %v; want ErrNotMajor", len(slots), m, err)
			}
		}
	}
}
