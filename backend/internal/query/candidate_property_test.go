package query_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
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

// In each of the twelve major keys, transposed from C, the candidates after
// these progressions keep the order of the groups, the degrees and the labels
// (all but the spelled chords), the same input gives the same output, and the
// labels are the facts of reading the progression with the candidate added
// (誤り 5 in docs/architecture/explanations.md), C7 always V7/IV (#83).
func TestNextCandidatesProperties(t *testing.T) {
	cat := transitions(t)
	c := progression.ChordOf(parse("C"))
	g := progression.ChordOf(parse("G"))
	inputs := []struct {
		p     progression.Progression
		beats int
	}{
		{after(), 4}, {after(), 16}, {after("C", "G7"), 4}, {after("Dm7"), 2}, {after("D7"), 4},
		{after("Fm"), 4}, {after("B♯"), 4}, {after("Gsus4"), 4}, {after("C6"), 4}, {after("C/E", "C7"), 4},
		{prog(major(note("C")), slot(c, 12), slot(g, 3)), 1},
		{prog(major(note("C")), slot(c, 8), slot(g, 1)), 1},
		{prog(major(note("C")), slot(g, 4), slot(progression.NoChord(), 4)), 4},
		{prog(major(note("C")), slot(g, 4), slot(progression.Unchosen(), 4)), 4},
	}
	for _, in := range inputs {
		inC := summary(candidates(t, in.p, in.beats))
		for _, tonic := range tonics {
			up, err := material.IntervalBetween(note("C"), note(tonic))
			if err != nil {
				t.Fatal(err)
			}
			p, err := in.p.Transpose(up)
			if errors.Is(err, material.ErrAlterOutOfRange) {
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			got := candidates(t, p, in.beats)
			if again := candidates(t, p, in.beats); !reflect.DeepEqual(got, again) {
				t.Errorf("%s major: two calls differ", tonic)
			}
			if s := summary(got); !slices.Equal(s, inC) {
				t.Errorf("%s major:\n got %v\nwant %v (C)", tonic, s, inC)
			}
			checkFacts(t, p, in.beats, got, cat)
		}
	}
}

// summary writes each candidate without its spelling: the group, the
// position, the reading and the labels.
func summary(groups []query.CandidateGroup) []string {
	var out []string
	for _, g := range groups {
		for _, c := range g.Candidates {
			numeral, _ := notation.FormatRoman(c.Position)
			m := c.Reading.Main
			parts := []string{fmt.Sprintf("%v %s %v %v %d %v", g.Kind, numeral, m.Kind, m.Function, m.Target, m.Modes)}
			if c.RootMotion != nil {
				parts = append(parts, c.RootMotion.Interval.String())
			}
			if c.CommonTones != nil {
				parts = append(parts, strconv.Itoa(c.CommonTones.Count))
			}
			for _, l := range []any{c.Resolution, c.Cadence, c.Preference} {
				if !reflect.ValueOf(l).IsNil() {
					parts = append(parts, fmt.Sprintf("%+v", l))
				}
			}
			s := strings.Join(parts, " ")
			out = append(out, s)
		}
	}
	return out
}

// checkFacts checks each candidate against the progression with it added.
func checkFacts(t *testing.T, p progression.Progression, beats int, groups []query.CandidateGroup, cat catalog.FunctionTransitions) {
	t.Helper()
	key := p.Key()
	for _, g := range groups {
		for _, c := range g.Candidates {
			next, err := p.Append(slot(progression.ChordOf(c.Chord), beats))
			if err != nil {
				t.Fatal(err)
			}
			read, err := reading.ReadProgression(next)
			if err != nil {
				t.Fatal(err)
			}
			n := next.Len()
			want, got := read[n-1].Chord.Main, c.Reading.Main
			name := fmt.Sprintf("%s major, %v", key.Tonic(), c.Chord.Notes())
			if got.Kind != g.Kind || got.Key != want.Key || got.Kind != want.Kind || got.Function != want.Function ||
				got.Target != want.Target || !slices.Equal(got.Modes, want.Modes) {
				t.Errorf("%s in group %v: reading %+v; ReadProgression %+v", name, g.Kind, got, want)
			}
			if target, _ := c.Position.Target(); target != got.Target {
				t.Errorf("%s: position points to %d; reading %d", name, target, got.Target)
			}
			if s, _ := notation.FormatChord(c.Chord); s == "C7" && key.Tonic() == note("C") && got.Target != 4 {
				t.Errorf("C7 in C: target %d; want 4 (V7/IV)", got.Target)
			}
			if !slices.Equal(c.Members.Notes, c.Chord.Notes()) {
				t.Errorf("%s: members %v", name, c.Members.Notes)
			}
			place := relation.Places(next)[n-1]
			if full := place.LastInPhrase && place.FullPhrase; full != (c.Cadence != nil) || full && c.Cadence.Cadence != read[n-1].Cadence {
				t.Errorf("%s: cadence %+v; ReadProgression %v, full phrase %v", name, c.Cadence, read[n-1].Cadence, full)
			}
			var lastChord material.Chord
			hasLast := false
			if n > 1 {
				lastChord, hasLast = p.Slots()[n-2].Content().Chord()
			}
			if hasLast != (c.CommonTones != nil) || hasLast != (c.Resolution != nil) || !hasLast && c.RootMotion != nil {
				t.Errorf("%s: relations %+v %+v %+v with a chord last %v", name, c.RootMotion, c.CommonTones, c.Resolution, hasLast)
			}
			if hasLast && c.Resolution.Resolution != read[n-2].Resolution {
				t.Errorf("%s: resolution %v; ReadProgression %v", name, c.Resolution.Resolution, read[n-2].Resolution)
			}
			if l := c.Preference; l != nil && (g.Kind != reading.Diatonic || l.Style == 0 || l.Style != cat.Source().Style || l.Basis() != query.Convention ||
				!hasLast || !reading.Tertian(lastChord) || l.From != read[n-2].Chord.Main.Function || l.To != got.Function) {
				t.Errorf("%s in group %v: preference %+v, basis %v", name, g.Kind, *l, l.Basis())
			}
		}
	}
}
