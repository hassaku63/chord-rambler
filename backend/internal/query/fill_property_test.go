package query_test

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// In each of the twelve major keys, transposed from C, the candidates for
// these slots keep the order and the labels (all but the spelled chords), the
// same input gives the same output, and the labels are the facts of reading
// the progression with the slot filled (誤り 5 in
// docs/architecture/explanations.md).
func TestSlotCandidatesProperties(t *testing.T) {
	none := progression.RoleNone
	inputs := []struct {
		p progression.Progression
		i int
	}{
		{inC(open(progression.RolePrepare), bar("G7")), 0},
		{inC(bar("G7"), open(progression.RolePrepare), bar("G7")), 1},
		{inC(bar("C"), open(progression.RolePrepare), open(none)), 1},
		{inC(bar("G7"), open(progression.RoleResolve)), 1},
		{inC(slot(progression.NoChord(), 4), open(progression.RoleResolve)), 1},
		{inC(bar("C"), bar("F"), bar("Dm"), bar("G7"), open(progression.RoleResolve)), 4},
		{inC(bar("C"), open(progression.RoleRise)), 1},
		{inC(bar("C"), bar("F"), bar("G7"), open(progression.RoleClose)), 3},
		{inC(bar("G7"), open(progression.RoleClose)), 1},
		{inC(bar("Gsus4"), open(progression.RoleLink), bar("C")), 1},
		{inC(bar("Am"), open(progression.RoleLink)), 1},
		{inC(bar("C"), open(progression.RoleStable)), 1},
		{inC(bar("C"), open(progression.RoleSpread)), 1},
		{inC(bar("C"), bar("F"), open(none), bar("C")), 2},
		{inC(bar("C"), bar("G7"), bar("C")), 1},
		{inC(bar("C6"), open(none), bar("Am")), 1},
	}
	for _, in := range inputs {
		inC := slotSummary(slotCandidates(t, in.p, in.i))
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
			got := slotCandidates(t, p, in.i)
			if again := slotCandidates(t, p, in.i); !reflect.DeepEqual(got, again) {
				t.Errorf("%s major: two calls differ", tonic)
			}
			if s := slotSummary(got); !slices.Equal(s, inC) {
				t.Errorf("%s major:\n got %v\nwant %v (C)", tonic, s, inC)
			}
			checkFillFacts(t, p, in.i, got)
		}
	}
}

// The candidates for a last slot with no role are those of NextCandidates
// for a slot of its length after the slots before it: the two share the
// range, the order and the labels.
func TestSlotCandidatesAtTheEndAreNextCandidates(t *testing.T) {
	c := progression.ChordOf(parse("C"))
	g := progression.ChordOf(parse("G"))
	inputs := []struct {
		p     progression.Progression
		beats int
	}{
		{after(), 4}, {after(), 16}, {after("C", "G7"), 4}, {after("Dm7"), 2}, {after("Gsus4"), 4},
		{prog(major(note("C")), slot(c, 12), slot(g, 3)), 1},
		{prog(major(note("C")), slot(g, 4), slot(progression.Unchosen(), 4)), 4},
	}
	for _, in := range inputs {
		want := candidates(t, in.p, in.beats)
		filled, err := in.p.Append(slot(progression.Unchosen(), in.beats))
		if err != nil {
			t.Fatal(err)
		}
		got := slotCandidates(t, filled, filled.Len()-1)
		if !reflect.DeepEqual(plainGroups(got), want) {
			t.Errorf("after %d slots, %d beats:\n got %v\nwant %v", in.p.Len(), in.beats, summary(plainGroups(got)), summary(want))
		}
	}
}

// plainGroups returns the groups with the labels of Candidate only.
func plainGroups(groups []query.SlotCandidateGroup) []query.CandidateGroup {
	out := make([]query.CandidateGroup, 0, len(groups))
	for _, g := range groups {
		cs := make([]query.Candidate, 0, len(g.Candidates))
		for _, c := range g.Candidates {
			cs = append(cs, c.Candidate)
		}
		out = append(out, query.CandidateGroup{Kind: g.Kind, Candidates: cs})
	}
	return out
}

// slotSummary writes each candidate as summary does, with the labels of the
// next slot and of the role.
func slotSummary(groups []query.SlotCandidateGroup) []string {
	out := summary(plainGroups(groups))
	n := 0
	for _, g := range groups {
		for _, c := range g.Candidates {
			out[n] += fmt.Sprintf(" %+v %+v %+v", c.NextResolution, c.NextCadence, c.Role)
			n++
		}
	}
	return out
}

// checkFillFacts checks each candidate against the progression with the slot
// filled by it.
func checkFillFacts(t *testing.T, p progression.Progression, i int, groups []query.SlotCandidateGroup) {
	t.Helper()
	n := 0
	for _, g := range groups {
		fitting := true
		for _, c := range g.Candidates {
			n++
			name := fmt.Sprintf("%s major, slot %d, %v", p.Key().Tonic(), i, c.Chord.Notes())
			filled, err := p.Fill(i, progression.ChordOf(c.Chord))
			if err != nil {
				t.Fatal(err)
			}
			read, err := reading.ReadProgression(filled)
			if err != nil {
				t.Fatal(err)
			}
			verdicts, err := reading.ReadRoleVerdicts(filled)
			if err != nil {
				t.Fatal(err)
			}
			places := relation.Places(filled)
			last := i+1 == filled.Len()
			if !reading.Tertian(c.Chord) {
				t.Errorf("%s: not a tertian chord", name)
			}
			if c.Reading.Main.Kind != g.Kind || !reflect.DeepEqual(c.Reading.Main, read[i].Chord.Main) {
				t.Errorf("%s in group %v: reading %+v; ReadProgression %+v", name, g.Kind, c.Reading.Main, read[i].Chord.Main)
			}
			if full := places[i].LastInPhrase && places[i].FullPhrase; full != (c.Cadence != nil) || full && c.Cadence.Cadence != read[i].Cadence {
				t.Errorf("%s: cadence %+v; ReadProgression %v, full phrase %v", name, c.Cadence, read[i].Cadence, full)
			}
			hasPrev := i > 0 && read[i-1].HasChord
			if hasPrev != (c.CommonTones != nil) || hasPrev != (c.Resolution != nil) || hasPrev && c.Resolution.Resolution != read[i-1].Resolution {
				t.Errorf("%s: relations %+v %+v with a chord before %v", name, c.CommonTones, c.Resolution, hasPrev)
			}
			if hasNext := !last && read[i+1].HasChord; hasNext != (c.NextResolution != nil) || hasNext && c.NextResolution.Resolution != read[i].Resolution {
				t.Errorf("%s: next resolution %+v; ReadProgression %v", name, c.NextResolution, read[i].Resolution)
			}
			if full := !last && places[i+1].LastInPhrase && places[i+1].FullPhrase; full != (c.NextCadence != nil) || full && c.NextCadence.Cadence != read[i+1].Cadence {
				t.Errorf("%s: next cadence %+v; ReadProgression %v", name, c.NextCadence, read[i+1].Cadence)
			}
			judged := verdicts[i].Fit != reading.RoleNotJudged
			if judged != (c.Role != nil) || judged && (c.Role.Fit != verdicts[i].Fit || !reflect.DeepEqual(c.Role.Unmet, verdicts[i].Unmet) ||
				c.Role.Role != p.Slots()[i].Role() || c.Role.Basis() != query.School) {
				t.Errorf("%s: role %+v; ReadRoleVerdicts %+v", name, c.Role, verdicts[i])
			}
			if l := c.Preference; l != nil && (g.Kind != reading.Diatonic || !hasPrev || l.From != read[i-1].Chord.Main.Function || l.To != c.Reading.Main.Function) {
				t.Errorf("%s: preference %+v", name, *l)
			}
			fit := c.Role != nil && c.Role.Fit == reading.RoleFits
			if fit && !fitting {
				t.Errorf("%s in group %v: fits after one that does not", name, g.Kind)
			}
			fitting = fitting && fit
		}
	}
	if n != 31 {
		t.Errorf("%d candidates; want 31", n)
	}
}
