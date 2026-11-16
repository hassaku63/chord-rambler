package query_test

import (
	"errors"
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
// the last chord is its reading (#102): Gsus4 and G7(♭9) are dominant, C6 and
// C/E tonic.
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
		{"Gsus4", after("Gsus4"), afterDominant},
		{"C6", prog(major(note("C")), slot(progression.ChordOf(chord("C", "M3 P5 M6", "")), 4)), afterTonic},
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
	}
	for _, tt := range tests {
		if _, err := query.NextCandidates(tt.p, tt.beats, transitions(t)); !errors.Is(err, tt.want) {
			t.Errorf("%s: error %v; want %v", tt.name, err, tt.want)
		}
	}
}
