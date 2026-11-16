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
	csus4   = chord(cNat, "P4 P5", "")
	dm      = chord(dNat, minTriad, "")
	d       = chord(dNat, maj, "")
	d7      = chord(dNat, dom7, "")
	e       = chord(eNat, maj, "")
	em      = chord(eNat, minTriad, "")
	f       = chord(fNat, maj, "")
	fm      = chord(fNat, minTriad, "")
	fsus2   = chord(fNat, "M2 P5", "")
	g       = chord(gNat, maj, "")
	g7      = chord(gNat, dom7, "")
	gOverB  = chord(gNat, maj, "M3")
	g7b9    = chord(gNat, "M3 P5 m7 m9", "")
	g7OverF = chord(gNat, dom7, "m7")
	g7n9    = chord(gNat, "M3 P5 m7 M9", "")
	gsus4   = chord(gNat, "P4 P5", "")
	g6      = chord(gNat, "M3 P5 M6", "")
	gm      = chord(gNat, minTriad, "")
	am      = chord(aNat, minTriad, "")
	a7      = chord(aNat, dom7, "")
	b7      = chord(bNat, dom7, "")
	bdim    = chord(bNat, dim, "")
	d6      = chord(dNat, "M3 P5 M6", "")
	c6      = chord(cNat, "M3 P5 M6", "")
	asus4   = chord(aNat, "P4 P5", "")
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

// bars is a slot of one bar for each chord.
func bars(chords ...material.Chord) []slot {
	var out []slot
	for _, ch := range chords {
		out = append(out, bar(ch))
	}
	return out
}

// How a dominant goes on, read on every slot of progressions in C major
// (docs/architecture/theory-conventions.md, 2; Hutchinson 17.4,
// https://musictheory.pugetsound.edu/mt21c/AnalyzingSecondaryDominants.html,
// and Hutchinson on irregular resolutions,
// https://musictheory.pugetsound.edu/mt21c/IrregularResoltionsOfSecondaryDominants.html).
// The owner's calls on #103: the next chord is expected when it is read as
// diatonic, is on the degree pointed to, and is a triad with a third or a
// seventh chord (G7, G/B and G7(♭9) after D7; not Gm, borrowed, nor D7 after
// A7, a secondary dominant). V is the major triad or the dominant seventh on
// the fifth degree, whatever its bass and tensions: G/B, G7/F and G7(9) are V;
// Gsus4 and G6 are not. Nothing is said before a sus or a sixth chord, when no
// chord follows, or when the same dominant goes on: the last of them says it.
func TestResolution(t *testing.T) {
	const (
		none      = reading.NoResolution
		expected  = reading.ResolutionExpected
		irregular = reading.ResolutionIrregular
	)
	tests := []struct {
		name  string
		slots []slot
		want  []reading.Resolution
	}{
		// The cases of #103.
		{"D7 → G", bars(d7, g), []reading.Resolution{expected, none}},
		{"D7 → Em", bars(d7, em), []reading.Resolution{irregular, none}},
		{"G7 → C", bars(g7, c), []reading.Resolution{expected, none}},
		{"G7 → Am", bars(g7, am), []reading.Resolution{irregular, none}},
		// The owner's calls on #103.
		{"D7 → G7", bars(d7, g7), []reading.Resolution{expected, none}},
		{"D7 → G/B", bars(d7, gOverB), []reading.Resolution{expected, none}},
		{"D7 → G7(♭9)", bars(d7, g7b9), []reading.Resolution{expected, none}},
		{"D7 → Gm", bars(d7, gm), []reading.Resolution{irregular, none}},
		{"A7 → D7 → G", bars(a7, d7, g), []reading.Resolution{irregular, expected, none}},
		{"G/B → C", bars(gOverB, c), []reading.Resolution{expected, none}},
		{"G7/F → C", bars(g7OverF, c), []reading.Resolution{expected, none}},
		{"G7(9) → C", bars(g7n9, c), []reading.Resolution{expected, none}},
		{"Gsus4 → C: not V", bars(gsus4, c), []reading.Resolution{none, none}},
		{"G6 → C: not V", bars(g6, c), []reading.Resolution{none, none}},
		{"G → Gsus4", bars(g, gsus4), []reading.Resolution{none, none}},
		{"G7 → Csus4", bars(g7, csus4), []reading.Resolution{none, none}},
		{"G → G7 → C", bars(g, g7, c), []reading.Resolution{none, expected, none}},
		{"G → G/B → C", bars(g, gOverB, c), []reading.Resolution{none, expected, none}},
		{"D7 → D7 → G", bars(d7, d7, g), []reading.Resolution{none, expected, none}},
		// What follows from the calls.
		{"D → D7 → G: the same target", bars(d, d7, g), []reading.Resolution{none, expected, none}},
		// A sus or a sixth chord has not the shape of a secondary dominant
		// (the owner's call on #103), so it says nothing, and what follows a
		// dominant that is not followed by a chord with a third says nothing.
		{"D6 → G: not a secondary dominant", bars(d6, g), []reading.Resolution{none, none}},
		{"G7 → Bdim", bars(g7, bdim), []reading.Resolution{irregular, none}},
		{"G7 → Asus4: a sus chord", bars(g7, asus4), []reading.Resolution{none, none}},
		{"D7 → C6: a sixth chord", bars(d7, c6), []reading.Resolution{none, none}},
		{"D7 → G6: a sixth chord", bars(d7, g6), []reading.Resolution{none, none}},
		{"A7 → Dm", bars(a7, dm), []reading.Resolution{expected, none}},
		{"B7 → Em", bars(b7, em), []reading.Resolution{expected, none}},
		{"C7 → F", bars(c7, f), []reading.Resolution{expected, none}},
		{"E → Am", bars(e, am), []reading.Resolution{expected, none}},
		{"G7 → Cm: borrowed", bars(g7, cm), []reading.Resolution{irregular, none}},
		{"G7 → C7: a secondary dominant", bars(g7, c7), []reading.Resolution{irregular, none}},
		{"C → F: not a dominant", bars(c, f), []reading.Resolution{none, none}},
		{"Am → Dm: not a dominant", bars(am, dm), []reading.Resolution{none, none}},
		{"D7 last", bars(d7), []reading.Resolution{none}},
		{"G7 → no chord", []slot{bar(g7), noChord}, []reading.Resolution{none, none}},
		{"D7 → no chord", []slot{bar(d7), noChord}, []reading.Resolution{none, none}},
		{"D7 → none chosen", []slot{bar(d7), unchosen}, []reading.Resolution{none, none}},
		// The next slot may be in the next phrase (decided in #103, not by the
		// owner).
		{"G7 | C", bars(c, f, dm, g7, c), []reading.Resolution{none, none, none, expected, none}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []reading.Resolution
			for _, r := range readProgression(t, build(major(cNat), tt.slots...)) {
				got = append(got, r.Resolution)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("resolutions = %v; want %v", got, tt.want)
			}
		})
	}
}

// The cadences of every slot of progressions in C major. A cadence is named
// on the last slot of a full phrase of four bars only
// (docs/architecture/theory-conventions.md, 4; Hutchinson,
// https://musictheory.pugetsound.edu/mt21c/cadences.html): authentic V → I,
// half on V, plagal IV → I, deceptive V → other than I. V, I and IV are read
// as in TestResolution (the owner's calls on #103): G7 → Cm is deceptive, and
// nothing is named before or on a sus or a sixth chord. A phrase still being
// written at the end gets none. A phrase ending on V is a half cadence
// whatever comes before it; the other cadences need a chord in the slot
// before, in the same phrase (decided in #103, not by the owner).
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
		{"authentic: G7 → C", bars(c, f, g7, c), []reading.Cadence{none, none, none, authentic}},
		{"half: ends on G", bars(c, am, f, g), []reading.Cadence{none, none, none, half}},
		{"plagal: F → C", bars(c, am, f, c), []reading.Cadence{none, none, none, plagal}},
		{"deceptive: G → Am", bars(c, f, g, am), []reading.Cadence{none, none, none, deceptive}},
		{"G7 → C in the middle of the phrase", bars(c, g7, c, am), []reading.Cadence{none, none, none, none}},
		// The owner's calls on #103.
		{"G7 → Cmaj7", bars(c, f, g7, cMaj7), []reading.Cadence{none, none, none, authentic}},
		{"G7 → C/E", bars(c, f, g7, cOverE), []reading.Cadence{none, none, none, authentic}},
		{"G7 → Cm", bars(c, f, g7, cm), []reading.Cadence{none, none, none, deceptive}},
		{"G/B → C", bars(c, f, gOverB, c), []reading.Cadence{none, none, none, authentic}},
		{"G7/F → C", bars(c, f, g7OverF, c), []reading.Cadence{none, none, none, authentic}},
		{"G7(9) → C", bars(c, f, g7n9, c), []reading.Cadence{none, none, none, authentic}},
		{"Gsus4 → C: not V", bars(c, f, gsus4, c), []reading.Cadence{none, none, none, none}},
		{"ends on Gsus4", bars(c, f, g, gsus4), []reading.Cadence{none, none, none, none}},
		{"G7 → Csus4", bars(c, f, g7, csus4), []reading.Cadence{none, none, none, none}},
		{"G7 → Bdim: the end of a full phrase", bars(c, f, g7, bdim), []reading.Cadence{none, none, none, deceptive}},
		{"D6 → G: ends on V", bars(c, f, d6, g), []reading.Cadence{none, none, none, half}},
		{"ends on G6: not V", bars(c, f, dm, g6), []reading.Cadence{none, none, none, none}},
		{"Fsus2 → C: not IV", bars(c, am, fsus2, c), []reading.Cadence{none, none, none, none}},
		{"Fm → C: borrowed", bars(c, am, fm, c), []reading.Cadence{none, none, none, none}},
		{"F | G7 | C: three bars", bars(f, g7, c), []reading.Cadence{none, none, none}},
		{
			"a full phrase, then three bars",
			bars(c, f, g7, c, f, g7, c),
			[]reading.Cadence{none, none, none, authentic, none, none, none},
		},
		// What follows from the calls.
		{"ends on G → G7: V", bars(c, f, g, g7), []reading.Cadence{none, none, none, half}},
		{"G7 → C7: a secondary dominant is not I", bars(c, f, g7, c7), []reading.Cadence{none, none, none, deceptive}},
		{"D7 → G: ends on V", bars(c, f, d7, g), []reading.Cadence{none, none, none, half}},
		{"Dm → C: not IV", bars(c, f, dm, c), []reading.Cadence{none, none, none, none}},
		{"one slot of four bars, G", []slot{{progression.ChordOf(g), 16}}, []reading.Cadence{half}},
		{
			"G and C of two beats end the phrase",
			[]slot{bar(c), bar(f), bar(dm), {progression.ChordOf(g), 2}, {progression.ChordOf(c), 2}},
			[]reading.Cadence{none, none, none, none, authentic},
		},
		{
			"each phrase ends on its own",
			bars(c, f, g7, c, am, dm, g7, am),
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
			var got []reading.Cadence
			for _, r := range readProgression(t, build(major(cNat), tt.slots...)) {
				got = append(got, r.Cadence)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("cadences = %v; want %v", got, tt.want)
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
