package query

import (
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// SlotCandidateGroup is one group of the candidates for a slot to fill: the
// chords of one kind of reading, in order.
type SlotCandidateGroup struct {
	// Kind is the kind every candidate of the group is read as: Diatonic,
	// SecondaryDominant or Borrowed.
	Kind       reading.Kind
	Candidates []SlotCandidate
}

// SlotCandidate is one chord that may fill a slot of a progression, and its
// labels. The labels of Candidate are of the slot filled: its reading and its
// cadence, and the relations from the slot before it, read as the last slot
// of NextCandidates. The others look at the slot after it and at the role of
// the slot.
type SlotCandidate struct {
	Candidate
	// NextResolution is how the candidate goes on to the next slot. It is
	// given when the next slot holds a chord.
	NextResolution *ResolutionLabel
	// NextCadence is the cadence of the next slot. It is given when the next
	// slot is the last of a full phrase.
	NextCadence *CadenceLabel
	Role        *RoleLabel
}

// RoleLabel holds whether the candidate fits the structural role of the slot,
// and the conditions of the role it does not meet, as reading.ReadRoleVerdicts
// reads the progression with the slot filled. It is given when the slot has a
// structural role, and may say RoleUndecided.
type RoleLabel struct {
	Role progression.Role
	Fit  reading.RoleFit
	// Unmet is set when Fit is RoleMisfits (reading.RoleVerdict).
	Unmet []reading.RoleCondition
}

// Basis returns School.
func (RoleLabel) Basis() Basis { return School }

// SlotCandidates returns the candidates for the slot at index i (from 0) of
// the progression: the chords of NextCandidates, in the same three groups. A
// candidate replaces only the content of the slot (progression.Fill), so the
// slot keeps its length and its role.
//
// In every group, the candidates that fit the structural role of the slot
// come first, and the rest (those that do not fit and those undecided) after
// them. Each of the two keeps the order of NextCandidates: the diatonic group
// by the preference from the function of the slot before, on the same
// conditions, and the other groups by the degree. A slot with no role or a
// colour role is ordered as NextCandidates orders it.
//
// Every candidate carries the labels of NextCandidates, with the slot before
// in place of the last slot; how it goes on to the next slot when that holds
// a chord; the cadence of the next slot when it is the last of a full phrase;
// and whether it fits the role, with the conditions not met, when the role is
// structural. The readings come from reading.ReadProgression and
// reading.ReadRoleVerdicts of the progression with the slot filled.
//
// It fails for a key other than Ionian (relation.ErrNotMajor), an index
// outside 0 to Len()-1 (progression.ErrIndex), and a candidate beyond a double
// sharp or a double flat (material.ErrAlterOutOfRange).
func SlotCandidates(p progression.Progression, i int, transitions catalog.FunctionTransitions) ([]SlotCandidateGroup, error) {
	key := p.Key()
	if key.Mode() != material.Ionian {
		return nil, fmt.Errorf("slot candidates in %s %s: %w", key.Tonic(), key.Mode(), relation.ErrNotMajor)
	}
	if i < 0 || i >= p.Len() {
		return nil, fmt.Errorf("slot candidates for slot %d of %d: %w", i, p.Len(), progression.ErrIndex)
	}
	read, err := reading.ReadProgression(p)
	if err != nil {
		return nil, fmt.Errorf("slot candidates: %w", err)
	}
	from := functionBefore(p, read, i)
	out := make([]SlotCandidateGroup, 0, len(candidateGroups))
	for _, g := range candidateGroups {
		cs := make([]SlotCandidate, 0, len(g.positions))
		for _, pos := range g.positions {
			c, err := slotCandidate(p, i, pos)
			if err != nil {
				return nil, err
			}
			cs = append(cs, c)
		}
		if g.kind == reading.Diatonic {
			cs = byPreference(cs, func(c *SlotCandidate) *Candidate { return &c.Candidate }, from, transitions)
		}
		out = append(out, SlotCandidateGroup{Kind: g.kind, Candidates: fitsFirst(cs)})
	}
	return out, nil
}

// fitsFirst returns the candidates that fit the role, then the rest, each in
// the order given.
func fitsFirst(cs []SlotCandidate) []SlotCandidate {
	var fits, rest []SlotCandidate
	for _, c := range cs {
		if c.Role != nil && c.Role.Fit == reading.RoleFits {
			fits = append(fits, c)
		} else {
			rest = append(rest, c)
		}
	}
	return append(fits, rest...)
}

// slotCandidate returns the candidate of the position in the slot at i of p,
// with its labels but the preference.
func slotCandidate(p progression.Progression, i int, pos material.Position) (SlotCandidate, error) {
	c, err := spell(p.Key(), pos)
	if err != nil {
		return SlotCandidate{}, err
	}
	next, err := p.Fill(i, progression.ChordOf(c))
	if err != nil {
		return SlotCandidate{}, fmt.Errorf("slot candidate: %w", err)
	}
	read, err := reading.ReadProgression(next)
	if err != nil {
		return SlotCandidate{}, fmt.Errorf("slot candidate: %w", err)
	}
	verdicts, err := reading.ReadRoleVerdicts(next)
	if err != nil {
		return SlotCandidate{}, fmt.Errorf("slot candidate: %w", err)
	}
	base, err := labelled(next, read, i, pos, c)
	if err != nil {
		return SlotCandidate{}, err
	}
	out := SlotCandidate{Candidate: base}
	if v := verdicts[i]; v.Fit != reading.RoleNotJudged {
		out.Role = &RoleLabel{Role: next.Slots()[i].Role(), Fit: v.Fit, Unmet: v.Unmet}
	}
	if i+1 == next.Len() {
		return out, nil
	}
	if read[i+1].HasChord {
		out.NextResolution = &ResolutionLabel{Resolution: read[i].Resolution}
	}
	if place := relation.Places(next)[i+1]; place.LastInPhrase && place.FullPhrase {
		out.NextCadence = &CadenceLabel{Cadence: read[i+1].Cadence}
	}
	return out, nil
}
