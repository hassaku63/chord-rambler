package query_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// bar is a bar of the chord with no role; open is a bar none chosen yet with
// the role; inC is the progression in C of the slots.
func bar(s string) progression.Slot { return slot(progression.ChordOf(parse(s)), 4) }

func open(r progression.Role) progression.Slot {
	s, err := slot(progression.Unchosen(), 4).WithRole(r)
	if err != nil {
		panic(err)
	}
	return s
}

func inC(slots ...progression.Slot) progression.Progression { return prog(major(note("C")), slots...) }

func slotCandidates(t *testing.T, p progression.Progression, i int) []query.SlotCandidateGroup {
	t.Helper()
	got, err := query.SlotCandidates(p, i, transitions(t))
	if err != nil {
		t.Fatalf("SlotCandidates: %v", err)
	}
	return got
}

// marked writes the candidates as romans does, with the role label after
// each: "+" fits, "-" misfits, "?" undecided, nothing for no label.
func marked(cs []query.SlotCandidate) string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		s := romans([]query.Candidate{c.Candidate})
		if c.Role != nil {
			s += map[reading.RoleFit]string{reading.RoleFits: "+", reading.RoleMisfits: "-", reading.RoleUndecided: "?"}[c.Role.Fit]
		}
		out = append(out, s)
	}
	return strings.Join(out, " ")
}

const (
	secondaryMisfit = "V/ii- V7/ii- V/iii- V7/iii- V7/IV- V/V- V7/V- V/vi- V7/vi-"
	borrowedMisfit  = "♭III- ♭IIImaj7- iv- iv7- ♭VI- ♭VImaj7- ♭VII- ♭VII7-"
	secondaryUndec  = "V/ii? V7/ii? V/iii? V7/iii? V7/IV? V/V? V7/V? V/vi? V7/vi?"
	borrowedUndec   = "♭III? ♭IIImaj7? iv? iv7? ♭VI? ♭VImaj7? ♭VII? ♭VII7?"
)

// The candidates that fit a structural role come first in each group, each
// part in the order of NextCandidates (the preference after a tertian chord,
// the degree otherwise). The verdicts are those of reading.ReadRoleFits
// (docs/architecture/theory-conventions.md, 7). A slot with no role or a
// colour role, and one whose verdict is undecided, keep the order of
// NextCandidates.
func TestSlotCandidatesOrder(t *testing.T) {
	const (
		prepare = progression.RolePrepare
		resolve = progression.RoleResolve
		rise    = progression.RoleRise
		closing = progression.RoleClose
		none    = progression.RoleNone
	)
	tests := []struct {
		name                          string
		p                             progression.Progression
		i                             int
		diatonic, secondary, borrowed string
	}{
		{"準備 first, before G7", inC(open(prepare), bar("G7")), 0,
			"ii+ ii7+ IV+ IVmaj7+ I- Imaj7- iii- iii7- V- V7- vi- vi7- vii°- viiø7-", secondaryMisfit, borrowedMisfit},
		{"準備 between G7 and G7", inC(bar("G7"), open(prepare), bar("G7")), 1,
			"ii:L+ ii7:L+ IV:L+ IVmaj7:L+ I:C- Imaj7:C- iii:C- iii7:C- vi:C- vi7:C- V- V7- vii°- viiø7-", secondaryMisfit, borrowedMisfit},
		{"準備 before none chosen", inC(bar("C"), open(prepare), open(none)), 1,
			"ii:C? ii7:C? IV:C? IVmaj7:C? I? Imaj7? iii? iii7? V:L? V7:L? vi? vi7? vii°:L? viiø7:L?", secondaryUndec, borrowedUndec},
		{"no role, before none chosen", inC(bar("C"), open(none), open(none)), 1,
			"ii:C ii7:C IV:C IVmaj7:C I Imaj7 iii iii7 V:L V7:L vi vi7 vii°:L viiø7:L", secondary, borrowed},
		{"解決 after G7", inC(bar("G7"), open(resolve)), 1,
			"I:C+ Imaj7:C+ iii:C- iii7:C- vi:C- vi7:C- ii:L- ii7:L- IV:L- IVmaj7:L- V- V7- vii°- viiø7-", secondaryMisfit, borrowedMisfit},
		{"解決 after no chord", inC(slot(progression.NoChord(), 4), open(resolve)), 1,
			"I? Imaj7? ii? ii7? iii? iii7? IV? IVmaj7? V? V7? vi? vi7? vii°? viiø7?", secondaryUndec, borrowedUndec},
		{"高まり after C", inC(bar("C"), open(rise)), 1,
			"V:L+ V7:L+ vii°:L+ viiø7:L+ ii:C- ii7:C- IV:C- IVmaj7:C- I- Imaj7- iii- iii7- vi- vi7-", secondaryMisfit, borrowedMisfit},
		// After V, every tertian chord on the last slot of a full phrase gets
		// a cadence: authentic, half or deceptive.
		{"締め after G7, last of a full phrase", inC(bar("C"), bar("F"), bar("G7"), open(closing)), 3,
			"I:C+ Imaj7:C+ iii:C+ iii7:C+ vi:C+ vi7:C+ ii:L+ ii7:L+ IV:L+ IVmaj7:L+ V+ V7+ vii°+ viiø7+",
			"V/ii+ V7/ii+ V/iii+ V7/iii+ V7/IV+ V/V+ V7/V+ V/vi+ V7/vi+",
			"♭III+ ♭IIImaj7+ iv+ iv7+ ♭VI+ ♭VImaj7+ ♭VII+ ♭VII7+"},
		{"締め after G7, last of a phrase not full", inC(bar("G7"), open(closing)), 1,
			"I:C? Imaj7:C? iii:C? iii7:C? vi:C? vi7:C? ii:L? ii7:L? IV:L? IVmaj7:L? V? V7? vii°? viiø7?", secondaryUndec, borrowedUndec},
		{"広がり (a colour role) after C", inC(bar("C"), open(progression.RoleSpread)), 1,
			"ii:C ii7:C IV:C IVmaj7:C I Imaj7 iii iii7 V:L V7:L vi vi7 vii°:L viiø7:L", secondary, borrowed},
		{"first, no role", inC(open(none), bar("G7")), 0, inDegreeOrder, secondary, borrowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slotCandidates(t, tt.p, tt.i)
			kinds := []reading.Kind{reading.Diatonic, reading.SecondaryDominant, reading.Borrowed}
			for k, want := range []string{tt.diatonic, tt.secondary, tt.borrowed} {
				if got[k].Kind != kinds[k] {
					t.Errorf("group %d is %v; want %v", k, got[k].Kind, kinds[k])
				}
				if s := marked(got[k].Candidates); s != want {
					t.Errorf("%v:\n got %s\nwant %s", kinds[k], s, want)
				}
			}
		})
	}
}

// labels writes the labels of a candidate but its members, reading and
// preference.
func labels(c query.SlotCandidate) string {
	var out []string
	if c.RootMotion != nil {
		out = append(out, "motion:"+c.RootMotion.Interval.String())
	}
	if c.CommonTones != nil {
		out = append(out, fmt.Sprintf("common:%d", c.CommonTones.Count))
	}
	if c.Resolution != nil {
		out = append(out, "res:"+c.Resolution.Resolution.String())
	}
	if c.Cadence != nil {
		out = append(out, "cad:"+c.Cadence.Cadence.String())
	}
	if c.NextResolution != nil {
		out = append(out, "next:"+c.NextResolution.Resolution.String())
	}
	if c.NextCadence != nil {
		out = append(out, "nextcad:"+c.NextCadence.Cadence.String())
	}
	if c.Role != nil {
		role := "role:" + c.Role.Role.String() + "/" + c.Role.Fit.String()
		if c.Role.Unmet != nil {
			role += fmt.Sprint(c.Role.Unmet)
		}
		out = append(out, role)
	}
	return strings.Join(out, " ")
}

// findSlot returns the candidate of the numeral.
func findSlot(t *testing.T, groups []query.SlotCandidateGroup, numeral string) query.SlotCandidate {
	t.Helper()
	for _, g := range groups {
		for _, c := range g.Candidates {
			if romans([]query.Candidate{c.Candidate}) == numeral {
				return c
			}
		}
	}
	t.Fatalf("no candidate %s", numeral)
	return query.SlotCandidate{}
}

// The labels from the slot before, to the slot after and of the role. The
// next slot's resolution needs a chord there; its cadence needs it to be the
// last of a full phrase (four bars).
func TestSlotCandidatesLabels(t *testing.T) {
	none := progression.RoleNone
	tests := []struct {
		name      string
		p         progression.Progression
		i         int
		candidate string // a numeral as romans writes it, with its preference
		want      string
	}{
		{"V7 before C", inC(open(none), bar("C")), 0, "V7", "next:ResolutionExpected"},
		{"V7 before Am", inC(open(none), bar("Am")), 0, "V7", "next:ResolutionIrregular"},
		{"V7 between C and C", inC(bar("C"), open(none), bar("C")), 1, "V7:L",
			"motion:P5 common:1 res:NoResolution next:ResolutionExpected"},
		{"I last of a full phrase", inC(bar("C"), bar("F"), bar("G7"), open(none)), 3, "I:C",
			"motion:P4 common:1 res:ResolutionExpected cad:AuthenticCadence"},
		{"V7 before C ending a full phrase", inC(bar("C"), bar("F"), open(none), bar("C")), 2, "V7:C",
			"motion:M2 common:1 res:NoResolution next:ResolutionExpected nextcad:AuthenticCadence"},
		{"IV before C ending a full phrase", inC(bar("C"), bar("F"), open(none), bar("C")), 2, "IV",
			"motion:P1 common:3 res:NoResolution next:NoResolution nextcad:PlagalCadence"},
		{"ii before C ending a full phrase", inC(bar("C"), bar("F"), open(none), bar("C")), 2, "ii",
			"motion:M6 common:2 res:NoResolution next:NoResolution nextcad:NoCadence"},
		{"V7 before none chosen ending a full phrase", inC(bar("C"), bar("F"), open(none), open(none)), 2, "V7:C",
			"motion:M2 common:1 res:NoResolution nextcad:NoCadence"},
		{"V7 before none chosen", inC(bar("C"), open(none), open(none)), 1, "V7:L",
			"motion:P5 common:1 res:NoResolution"},
		// D7 resolves as expected to V, which does not fit 解決: D7 is not V.
		{"解決: V after D7", inC(bar("D7"), open(progression.RoleResolve)), 1, "V",
			"motion:P4 common:1 res:ResolutionExpected role:RoleResolve/RoleMisfits[ConditionPreviousV]"},
		{"解決: vi after G7", inC(bar("G7"), open(progression.RoleResolve)), 1, "vi:C",
			"motion:M2 common:0 res:ResolutionIrregular role:RoleResolve/RoleMisfits[ConditionExpectedResolution]"},
		{"解決: I after G7", inC(bar("G7"), open(progression.RoleResolve)), 1, "I:C",
			"motion:P4 common:1 res:ResolutionExpected role:RoleResolve/RoleFits"},
		{"高まり: I after C", inC(bar("C"), open(progression.RoleRise)), 1, "I",
			"motion:P1 common:3 res:NoResolution role:RoleRise/RoleMisfits[ConditionFunction]"},
		{"準備: I before C", inC(open(progression.RolePrepare), bar("C")), 0, "I",
			"next:NoResolution role:RolePrepare/RoleMisfits[ConditionFunction ConditionNextDominant]"},
		{"準備: IV before G7", inC(open(progression.RolePrepare), bar("G7")), 0, "IV",
			"next:NoResolution role:RolePrepare/RoleFits"},
		{"締め: I after G7", inC(bar("C"), bar("F"), bar("G7"), open(progression.RoleClose)), 3, "I:C",
			"motion:P4 common:1 res:ResolutionExpected cad:AuthenticCadence role:RoleClose/RoleFits"},
		{"締め: I after G7, phrase not full", inC(bar("F"), bar("G7"), open(progression.RoleClose)), 2, "I:C",
			"motion:P4 common:1 res:ResolutionExpected role:RoleClose/RoleUndecided"},
	}
	for _, tt := range tests {
		if got := labels(findSlot(t, slotCandidates(t, tt.p, tt.i), tt.candidate)); got != tt.want {
			t.Errorf("%s:\n got %s\nwant %s", tt.name, got, tt.want)
		}
	}
}

func TestSlotCandidatesErrors(t *testing.T) {
	minor, err := material.NewScale(note("A"), material.Aeolian)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		p    progression.Progression
		i    int
		want error
	}{
		{"a minor key", prog(minor, bar("Am")), 0, relation.ErrNotMajor},
		{"before the first slot", inC(bar("C")), -1, progression.ErrIndex},
		{"after the last slot", inC(bar("C")), 1, progression.ErrIndex},
		{"no slots", inC(), 0, progression.ErrIndex},
		{"a candidate not spelled", prog(major(note("E♯")), slot(progression.Unchosen(), 4)), 0, material.ErrAlterOutOfRange},
	}
	for _, tt := range tests {
		if _, err := query.SlotCandidates(tt.p, tt.i, transitions(t)); !errors.Is(err, tt.want) {
			t.Errorf("%s: error %v; want %v", tt.name, err, tt.want)
		}
	}
}
