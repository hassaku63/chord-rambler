package query_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

func parse(s string) material.Chord {
	c, err := notation.ParseChord(s)
	if err != nil {
		panic(err)
	}
	return c
}

func slot(c progression.Content, beats int) progression.Slot {
	s, err := progression.NewSlot(c, beats)
	if err != nil {
		panic(err)
	}
	return s
}

// prog builds the progression in the key from the slots.
func prog(key material.Scale, slots ...progression.Slot) progression.Progression {
	p := progression.New(key, progression.Meter{})
	for _, s := range slots {
		var err error
		if p, err = p.Append(s); err != nil {
			panic(err)
		}
	}
	return p
}

// after is a progression in C of one bar of each chord.
func after(chords ...string) progression.Progression {
	var slots []progression.Slot
	for _, s := range chords {
		slots = append(slots, slot(progression.ChordOf(parse(s)), 4))
	}
	return prog(major(note("C")), slots...)
}

func transitions(t *testing.T) catalog.FunctionTransitions {
	t.Helper()
	c, err := catalog.LoadFunctionTransitions()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func candidates(t *testing.T, p progression.Progression, beats int) []query.CandidateGroup {
	t.Helper()
	got, err := query.NextCandidates(p, beats, transitions(t))
	if err != nil {
		t.Fatalf("NextCandidates: %v", err)
	}
	return got
}

// romans writes the candidates as Roman numerals, a preference label as ":C"
// (Common) or ":L" (LessCommon) after its numeral.
func romans(cs []query.Candidate) string {
	var out []string
	for _, c := range cs {
		s, err := notation.FormatRoman(c.Position)
		if err != nil {
			panic(err)
		}
		if l := c.Preference; l != nil {
			s += map[catalog.Preference]string{catalog.Common: ":C", catalog.LessCommon: ":L"}[l.Preference]
		}
		out = append(out, s)
	}
	return strings.Join(out, " ")
}

const (
	inDegreeOrder = "I Imaj7 ii ii7 iii iii7 IV IVmaj7 V V7 vi vi7 vii° viiø7"
	secondary     = "V/ii V7/ii V/iii V7/iii V7/IV V/V V7/V V/vi V7/vi"
	borrowed      = "♭III ♭IIImaj7 iv iv7 ♭VI ♭VImaj7 ♭VII ♭VII7"
)

// The diatonic group follows the catalogue (function_transitions.json, from
// Open Music Theory): T → S, S → D and D → T are common, T → D, S → T and
// D → S less common, and a move within a function unstated. The function of
// the last chord is its reading (#102): G7(♭9) is dominant, C/E tonic. The
// preference is used only after a tertian chord (#103, the owner's call on
// #116), so Gsus4 and C6 keep the order of the degrees.
func TestNextCandidatesOrder(t *testing.T) {
	afterTonic := "ii:C ii7:C IV:C IVmaj7:C I Imaj7 iii iii7 V:L V7:L vi vi7 vii°:L viiø7:L"
	afterDominant := "I:C Imaj7:C iii:C iii7:C vi:C vi7:C ii:L ii7:L IV:L IVmaj7:L V V7 vii° viiø7"
	tests := []struct {
		name     string
		p        progression.Progression
		diatonic string
	}{
		{"G7", after("C", "G7"), afterDominant},
		{"Dm7", after("Dm7"), "V:C V7:C vii°:C viiø7:C I:L Imaj7:L ii ii7 iii:L iii7:L IV IVmaj7 vi:L vi7:L"},
		{"C", after("C"), afterTonic},
		{"Am", after("Am"), afterTonic},
		{"D7, a secondary dominant", after("D7"), inDegreeOrder},
		{"Fm, borrowed", after("Fm"), inDegreeOrder},
		{"B♯, facts only", after("B♯"), inDegreeOrder},
		{"Gsus4", after("Gsus4"), inDegreeOrder},
		{"C6", after("C6"), inDegreeOrder},
		{"C/E", after("C/E"), afterTonic},
		{"G7(♭9)", after("G7(♭9)"), afterDominant},
		{"no chord", prog(major(note("C")), slot(progression.ChordOf(parse("G7")), 4), slot(progression.NoChord(), 4)), inDegreeOrder},
		{"none chosen yet", prog(major(note("C")), slot(progression.ChordOf(parse("G7")), 4), slot(progression.Unchosen(), 4)), inDegreeOrder},
		{"empty", after(), inDegreeOrder},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := candidates(t, tt.p, 4)
			kinds := []reading.Kind{reading.Diatonic, reading.SecondaryDominant, reading.Borrowed}
			for i, want := range []string{tt.diatonic, secondary, borrowed} {
				if got[i].Kind != kinds[i] {
					t.Errorf("group %d is %v; want %v", i, got[i].Kind, kinds[i])
				}
				if s := romans(got[i].Candidates); s != want {
					t.Errorf("%v:\n got %s\nwant %s", kinds[i], s, want)
				}
			}
		})
	}
}

// The preference label keeps the functions it is about: from the last chord
// to the candidate.
func TestNextCandidatesPreferenceFunctions(t *testing.T) {
	tests := []struct {
		last, candidate string
		want            query.PreferenceLabel
	}{
		{"G7", "I", query.PreferenceLabel{From: reading.Dominant, To: reading.Tonic, Preference: catalog.Common}},
		{"G7", "IV", query.PreferenceLabel{From: reading.Dominant, To: reading.Subdominant, Preference: catalog.LessCommon}},
		{"Dm7", "V7", query.PreferenceLabel{From: reading.Subdominant, To: reading.Dominant, Preference: catalog.Common}},
		{"C/E", "V", query.PreferenceLabel{From: reading.Tonic, To: reading.Dominant, Preference: catalog.LessCommon}},
	}
	for _, tt := range tests {
		tt.want.Style, tt.want.Grounds = catalog.Classical, catalog.Convention
		got := find(t, candidates(t, after(tt.last), 4), tt.candidate).Preference
		if got == nil || *got != tt.want {
			t.Errorf("%s then %s: preference %+v; want %+v", tt.last, tt.candidate, got, tt.want)
		}
	}
}

// With the zero value of the catalogue, which has no entries, the diatonic
// group keeps the order of the degrees and no candidate has a preference.
func TestNextCandidatesWithoutCatalogue(t *testing.T) {
	got, err := query.NextCandidates(after("C", "G7"), 4, catalog.FunctionTransitions{})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{inDegreeOrder, secondary, borrowed} {
		if s := romans(got[i].Candidates); s != want {
			t.Errorf("group %d:\n got %s\nwant %s", i, s, want)
		}
	}
}

// Each candidate in C, read with nothing before it: the kind, the function,
// the degree it points to and the modes it is borrowed from. A borrowed chord
// is held by every parallel mode that has its notes.
func TestNextCandidatesReadingsInC(t *testing.T) {
	const (
		dia = reading.Diatonic
		sec = reading.SecondaryDominant
		bor = reading.Borrowed
		T   = reading.Tonic
		S   = reading.Subdominant
		D   = reading.Dominant
		no  = reading.NoFunction
	)
	dor, phr, mix, aeo, loc := material.Dorian, material.Phrygian, material.Mixolydian, material.Aeolian, material.Locrian
	tests := []struct {
		numeral  string
		kind     reading.Kind
		function reading.Function
		target   int
		modes    []material.Mode
	}{
		{"I", dia, T, 0, nil}, {"Imaj7", dia, T, 0, nil}, {"ii", dia, S, 0, nil}, {"ii7", dia, S, 0, nil},
		{"iii", dia, T, 0, nil}, {"iii7", dia, T, 0, nil}, {"IV", dia, S, 0, nil}, {"IVmaj7", dia, S, 0, nil},
		{"V", dia, D, 0, nil}, {"V7", dia, D, 0, nil}, {"vi", dia, T, 0, nil}, {"vi7", dia, T, 0, nil},
		{"vii°", dia, D, 0, nil}, {"viiø7", dia, D, 0, nil},
		{"V/ii", sec, no, 2, nil}, {"V7/ii", sec, no, 2, nil}, {"V/iii", sec, no, 3, nil}, {"V7/iii", sec, no, 3, nil},
		{"V7/IV", sec, no, 4, nil}, {"V/V", sec, no, 5, nil}, {"V7/V", sec, no, 5, nil},
		{"V/vi", sec, no, 6, nil}, {"V7/vi", sec, no, 6, nil},
		{"♭III", bor, no, 0, []material.Mode{dor, phr, aeo}},
		{"♭IIImaj7", bor, no, 0, []material.Mode{dor, aeo}},
		{"iv", bor, no, 0, []material.Mode{phr, aeo, loc}},
		{"iv7", bor, no, 0, []material.Mode{phr, aeo, loc}},
		{"♭VI", bor, no, 0, []material.Mode{phr, aeo, loc}},
		{"♭VImaj7", bor, no, 0, []material.Mode{phr, aeo}},
		{"♭VII", bor, no, 0, []material.Mode{dor, mix, aeo}},
		{"♭VII7", bor, no, 0, []material.Mode{aeo}},
	}
	groups := candidates(t, after(), 4)
	if n := len(groups[0].Candidates) + len(groups[1].Candidates) + len(groups[2].Candidates); n != len(tests) {
		t.Errorf("%d candidates; want %d", n, len(tests))
	}
	for _, tt := range tests {
		m := find(t, groups, tt.numeral).Reading.Main
		if m.Kind != tt.kind || m.Function != tt.function || m.Target != tt.target || !slices.Equal(m.Modes, tt.modes) {
			t.Errorf("%s: %v %v %d %v; want %v %v %d %v",
				tt.numeral, m.Kind, m.Function, m.Target, m.Modes, tt.kind, tt.function, tt.target, tt.modes)
		}
	}
}

// The chords are spelled in the key from their positions: ♭VI on D♭ is B𝄫.
func TestNextCandidatesSpelling(t *testing.T) {
	tests := []struct{ tonic, diatonic, secondary, borrowed string }{
		{"C",
			"C Cmaj7 Dm Dm7 Em Em7 F Fmaj7 G G7 Am Am7 Bdim Bm7(♭5)",
			"A A7 B B7 C7 D D7 E E7",
			"E♭ E♭maj7 Fm Fm7 A♭ A♭maj7 B♭ B♭7"},
		{"D♭",
			"D♭ D♭maj7 E♭m E♭m7 Fm Fm7 G♭ G♭maj7 A♭ A♭7 B♭m B♭m7 Cdim Cm7(♭5)",
			"B♭ B♭7 C C7 D♭7 E♭ E♭7 F F7",
			"F♭ F♭maj7 G♭m G♭m7 B𝄫 B𝄫maj7 C♭ C♭7"},
	}
	for _, tt := range tests {
		got := candidates(t, prog(major(note(tt.tonic))), 4)
		for i, want := range []string{tt.diatonic, tt.secondary, tt.borrowed} {
			var s []string
			for _, c := range got[i].Candidates {
				n, err := notation.FormatChord(c.Chord)
				if err != nil {
					t.Fatal(err)
				}
				s = append(s, n)
			}
			if strings.Join(s, " ") != want {
				t.Errorf("%s major, group %d:\n got %s\nwant %s", tt.tonic, i, strings.Join(s, " "), want)
			}
		}
	}
}

// find returns the candidate of the numeral.
func find(t *testing.T, groups []query.CandidateGroup, numeral string) query.Candidate {
	t.Helper()
	for _, g := range groups {
		for _, c := range g.Candidates {
			if s, _ := notation.FormatRoman(c.Position); s == numeral {
				return c
			}
		}
	}
	t.Fatalf("no candidate %s", numeral)
	return query.Candidate{}
}

// The relations to the chord of the last slot. Root motion is the interval up
// from root to root; B♯ to F has none (a fifth of five half steps). Common
// tones are counted by pitch class: B♯ sounds as C.
func TestNextCandidatesRelations(t *testing.T) {
	tests := []struct {
		last, candidate string
		motion          string // "" for no label
		common          int
		resolution      reading.Resolution
	}{
		{"G", "V7", "P1", 3, reading.NoResolution}, // the same dominant goes on
		{"G7", "I", "P4", 1, reading.ResolutionExpected},
		{"G7", "vi", "M2", 0, reading.ResolutionIrregular},
		{"D7", "V", "P4", 1, reading.ResolutionExpected},
		{"D7", "V7/V", "P1", 4, reading.NoResolution},
		{"D7", "iii", "M2", 0, reading.ResolutionIrregular},
		{"Fm", "♭VII7", "P4", 2, reading.NoResolution},
		{"B♯", "iv", "", 1, reading.NoResolution},
		{"B♯", "I", "d2", 3, reading.NoResolution},
		{"Gsus4", "I", "P4", 2, reading.NoResolution},
	}
	for _, tt := range tests {
		c := find(t, candidates(t, after(tt.last), 4), tt.candidate)
		motion := ""
		if c.RootMotion != nil {
			motion = c.RootMotion.Interval.String()
		}
		if motion != tt.motion || c.CommonTones == nil || c.CommonTones.Count != tt.common ||
			c.Resolution == nil || c.Resolution.Resolution != tt.resolution {
			t.Errorf("%s then %s: motion %q, common %+v, resolution %+v; want %q, %d, %v",
				tt.last, tt.candidate, motion, c.CommonTones, c.Resolution, tt.motion, tt.common, tt.resolution)
		}
	}
}

// The cadence label is given when the candidate makes the phrase full (16
// beats in 4/4), and not in the middle of the third bar.
func TestNextCandidatesCadence(t *testing.T) {
	c := progression.ChordOf(parse("C"))
	g := progression.ChordOf(parse("G"))
	f := progression.ChordOf(parse("F"))
	tests := []struct {
		name      string
		p         progression.Progression
		beats     int
		candidate string
		want      *reading.Cadence // nil for no label
	}{
		{"V to I in beat 16", prog(major(note("C")), slot(c, 4), slot(f, 4), slot(c, 4), slot(g, 3)), 1, "I", ptr(reading.AuthenticCadence)},
		{"V to vi in beat 16", prog(major(note("C")), slot(c, 12), slot(g, 3)), 1, "vi", ptr(reading.DeceptiveCadence)},
		{"IV to I in beat 16", prog(major(note("C")), slot(c, 12), slot(f, 3)), 1, "I", ptr(reading.PlagalCadence)},
		{"V in beat 16", prog(major(note("C")), slot(c, 15)), 1, "V7", ptr(reading.HalfCadence)},
		{"ii in beat 16", prog(major(note("C")), slot(c, 15)), 1, "ii", ptr(reading.NoCadence)},
		{"a whole phrase from the start", prog(major(note("C"))), 16, "I", ptr(reading.NoCadence)},
		{"V to I in the third bar", prog(major(note("C")), slot(c, 8), slot(g, 1)), 1, "I", nil},
		{"V to I in the next phrase", prog(major(note("C")), slot(c, 12), slot(g, 4)), 4, "I", nil},
	}
	for _, tt := range tests {
		got := find(t, candidates(t, tt.p, tt.beats), tt.candidate).Cadence
		switch {
		case tt.want == nil && got != nil:
			t.Errorf("%s: cadence %v; want no label", tt.name, got.Cadence)
		case tt.want != nil && (got == nil || got.Cadence != *tt.want):
			t.Errorf("%s: cadence %+v; want %v", tt.name, got, *tt.want)
		}
	}
}

func ptr[T any](v T) *T { return &v }

func TestNextCandidatesErrors(t *testing.T) {
	minor, err := material.NewScale(note("A"), material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		p     progression.Progression
		beats int
		want  error
	}{
		{"a minor key", progression.New(minor, progression.Meter{}), 4, relation.ErrNotMajor},
		{"no beats", after("C"), 0, progression.ErrLength},
		{"across the phrase", prog(major(note("C")), slot(progression.ChordOf(parse("C")), 15)), 4, progression.ErrCrossesPhrase},
		// V/iii in E♯ is D𝄪, whose third is beyond a double sharp: the whole
		// query fails.
		{"a candidate not spelled", prog(major(note("E♯"))), 4, material.ErrAlterOutOfRange},
		// The farthest candidates, V/iii (F𝄪) in G♯ and ♭VI (D𝄫) in F♭, are
		// spelled.
		{"G♯ major", prog(major(note("G♯"))), 4, nil},
		{"F♭ major", prog(major(note("F♭"))), 4, nil},
	}
	for _, tt := range tests {
		got, err := query.NextCandidates(tt.p, tt.beats, transitions(t))
		if !errors.Is(err, tt.want) {
			t.Errorf("%s: error %v; want %v", tt.name, err, tt.want)
		}
		if n := 0; err == nil {
			for _, g := range got {
				n += len(g.Candidates)
			}
			if n != 31 {
				t.Errorf("%s: %d candidates; want 31", tt.name, n)
			}
		}
	}
}
