package query

import (
	"errors"
	"fmt"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/progression"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
	"github.com/hassaku63/chord-rambler/backend/internal/relation"
)

// CandidateGroup is one group of candidates (まとまり): the chords of one kind
// of reading, in order.
type CandidateGroup struct {
	// Kind is the kind every candidate of the group is read as: Diatonic,
	// SecondaryDominant or Borrowed.
	Kind       reading.Kind
	Candidates []Candidate
}

// Candidate is one chord that may follow the last slot of a progression, and
// its labels. A label that does not apply is nil; one that applies may hold
// the zero value of its fact, such as NoResolution.
type Candidate struct {
	Chord material.Chord
	// Position is the chord seen from the tonic: the degree, its alteration,
	// the shape and, for a secondary dominant, the degree it points to.
	Position material.Position
	Members  MembersLabel
	// Reading is the main reading of the candidate as the last slot of the
	// progression with it added.
	Reading     ReadingLabel
	RootMotion  *RootMotionLabel
	CommonTones *CommonTonesLabel
	Resolution  *ResolutionLabel
	Cadence     *CadenceLabel
	Preference  *PreferenceLabel
}

// RootMotionLabel holds the interval up from the root of the chord of the
// last slot to the root of the candidate (relation.RootMotion). It is given
// when the last slot holds a chord and the two roots have an interval.
type RootMotionLabel struct {
	Interval material.Interval
}

// Basis returns Computation.
func (RootMotionLabel) Basis() Basis { return Computation }

// CommonTonesLabel holds how many pitch classes sound in both the chord of
// the last slot and the candidate (relation.CommonTones). It is given when the
// last slot holds a chord.
type CommonTonesLabel struct {
	Count int
}

// Basis returns Computation.
func (CommonTonesLabel) Basis() Basis { return Computation }

// ResolutionLabel holds how a chord goes on to the next, as
// reading.ReadProgression reads the progression with the candidate placed: in
// Candidate, the chord of the slot before the candidate to the candidate,
// given when that slot holds a chord; in SlotCandidate.NextResolution, the
// candidate to the next slot.
type ResolutionLabel struct {
	Resolution reading.Resolution
}

// Basis returns School.
func (ResolutionLabel) Basis() Basis { return School }

// CadenceLabel holds the cadence a slot ends its phrase with, as
// reading.ReadProgression reads the progression with the candidate placed: in
// Candidate, the slot of the candidate, given when it is the last slot of a
// full phrase; in SlotCandidate.NextCadence, the next slot.
type CadenceLabel struct {
	Cadence reading.Cadence
}

// Basis returns School.
func (CadenceLabel) Basis() Basis { return School }

// PreferenceLabel holds how often the catalogue of function transitions says
// the function of the last slot (From) moves to the function of the candidate
// (To), the style the catalogue limits that to, and the grounds of the entry.
// The preference is of the move between the functions, not of the chord, so
// the label keeps both. It is given to a diatonic candidate when the
// preference orders the group and the entry is Common or LessCommon.
type PreferenceLabel struct {
	From       reading.Function
	To         reading.Function
	Preference catalog.Preference
	Style      catalog.Style
	Grounds    catalog.Grounds
}

// Basis returns the basis of the grounds of the entry: Convention.
func (l PreferenceLabel) Basis() Basis { return basisOfGrounds(l.Grounds) }

// candidateGroups is the range of candidates (the owner's call A1 on #112): 31
// chords in C, in the order of the groups and, in each group, of the degree,
// the triad before the seventh chord. A secondary dominant goes by the degree
// it points to; V/IV as a triad is I, so it is left out. The borrowed chords
// come from the parallel minor.
var candidateGroups = []struct {
	kind      reading.Kind
	positions []material.Position
}{
	{reading.Diatonic, positions(
		"I", "Imaj7", "ii", "ii7", "iii", "iii7", "IV", "IVmaj7",
		"V", "V7", "vi", "vi7", "vii°", "viiø7")},
	{reading.SecondaryDominant, positions(
		"V/ii", "V7/ii", "V/iii", "V7/iii", "V7/IV", "V/V", "V7/V", "V/vi", "V7/vi")},
	{reading.Borrowed, positions(
		"♭III", "♭IIImaj7", "iv", "iv7", "♭VI", "♭VImaj7", "♭VII", "♭VII7")},
}

func positions(numerals ...string) []material.Position {
	out := make([]material.Position, 0, len(numerals))
	for _, s := range numerals {
		p, err := notation.ParseRoman(s)
		if err != nil {
			panic(err) // the table above holds valid numerals only
		}
		out = append(out, p)
	}
	return out
}

// NextCandidates returns the candidates for a slot of the beats added after
// the last slot of the progression, in three groups in this order: diatonic,
// secondary dominants and borrowed chords. Every group holds all of its
// chords, spelled in the key from their positions.
//
// The beats are the length the caller will give the slot. A caller whose
// slot would cross the boundary of a phrase cuts it to the beats left in the
// phrase before it calls; the cadence label depends on the length.
//
// The diatonic group is ordered by the catalogue of function transitions:
// the candidates the function of the chord of the last slot moves to with the
// preference Common come first, and the rest after them. The order is
// otherwise the order of candidateGroups, which the other groups keep, and so
// does the diatonic group when the progression is empty, its last slot holds
// no chord or none chosen yet, the chord is not tertian (reading.Tertian:
// a sus or a sixth chord), the chord has no function, or the catalogue has no
// entry for it (the zero value).
//
// Every candidate carries its members and its reading; the relations to the
// last slot when it holds a chord; the cadence when it ends a full phrase; and
// the preference on the conditions of PreferenceLabel. The readings come from
// reading.ReadProgression of the progression with the candidate added.
//
// It fails for a key other than Ionian (relation.ErrNotMajor), a length of
// less than one beat (progression.ErrLength), a slot that would cross the
// boundary of a phrase (progression.ErrCrossesPhrase), and a candidate beyond
// a double sharp or a double flat (material.ErrAlterOutOfRange). A key with
// one such candidate, such as E♯ major, fails as a whole.
func NextCandidates(p progression.Progression, beats int, transitions catalog.FunctionTransitions) ([]CandidateGroup, error) {
	key := p.Key()
	if key.Mode() != material.Ionian {
		return nil, fmt.Errorf("candidates in %s %s: %w", key.Tonic(), key.Mode(), relation.ErrNotMajor)
	}
	read, err := reading.ReadProgression(p)
	if err != nil {
		return nil, fmt.Errorf("candidates: %w", err)
	}
	from := functionBefore(p, read, p.Len())
	out := make([]CandidateGroup, 0, len(candidateGroups))
	for _, g := range candidateGroups {
		cs := make([]Candidate, 0, len(g.positions))
		for _, pos := range g.positions {
			c, err := candidate(p, pos, beats)
			if err != nil {
				return nil, err
			}
			cs = append(cs, c)
		}
		if g.kind == reading.Diatonic {
			cs = byPreference(cs, func(c *Candidate) *Candidate { return c }, from, transitions)
		}
		out = append(out, CandidateGroup{Kind: g.kind, Candidates: cs})
	}
	return out, nil
}

// functionBefore returns the function the preference starts from for a
// candidate at slot i: that of the chord of the slot before it when the slot
// holds a tertian chord (reading.Tertian), and NoFunction otherwise or when
// there is no slot before it. read is the reading of p.
func functionBefore(p progression.Progression, read []reading.SlotReading, i int) reading.Function {
	if i == 0 {
		return reading.NoFunction
	}
	if c, ok := p.Slots()[i-1].Content().Chord(); ok && reading.Tertian(c) {
		return read[i-1].Chord.Main.Function
	}
	return reading.NoFunction
}

// byPreference returns the candidates ordered by the preference from the
// function: Common first, then the rest, each in the order given. It gives the
// preference label to those of Common and LessCommon. of returns the
// candidate's Candidate, where the label goes.
func byPreference[T any](cs []T, of func(*T) *Candidate, from reading.Function, transitions catalog.FunctionTransitions) []T {
	var common, rest []T
	for _, c := range cs {
		cand := of(&c)
		t, ok := transitions.Lookup(from, cand.Reading.Main.Function)
		if ok && (t.Preference == catalog.Common || t.Preference == catalog.LessCommon) {
			cand.Preference = &PreferenceLabel{
				From: t.From, To: t.To, Preference: t.Preference,
				Style: transitions.Source().Style, Grounds: t.Grounds,
			}
		}
		if ok && t.Preference == catalog.Common {
			common = append(common, c)
		} else {
			rest = append(rest, c)
		}
	}
	return append(common, rest...)
}

// candidate returns the candidate of the position added after the last slot
// of p, with its labels but the preference.
func candidate(p progression.Progression, pos material.Position, beats int) (Candidate, error) {
	c, err := spell(p.Key(), pos)
	if err != nil {
		return Candidate{}, err
	}
	slot, err := progression.NewSlot(progression.ChordOf(c), beats)
	if err != nil {
		return Candidate{}, fmt.Errorf("candidate: %w", err)
	}
	next, err := p.Append(slot)
	if err != nil {
		return Candidate{}, fmt.Errorf("candidate: %w", err)
	}
	read, err := reading.ReadProgression(next)
	if err != nil {
		return Candidate{}, fmt.Errorf("candidate: %w", err)
	}
	return labelled(next, read, next.Len()-1, pos, c)
}

// spell returns the chord of the position in the key.
func spell(key material.Scale, pos material.Position) (material.Chord, error) {
	c, err := pos.Chord(key.Tonic())
	if err != nil {
		return material.Chord{}, fmt.Errorf("candidate in %s major: %w", key.Tonic(), err)
	}
	return c, nil
}

// labelled returns the candidate c, of the position, at slot i of next (the
// progression with the candidate placed, read as read) with its labels but the
// preference: the members, the reading, the cadence when slot i ends a full
// phrase, and the relations to the slot before when it holds a chord.
func labelled(next progression.Progression, read []reading.SlotReading, i int, pos material.Position, c material.Chord) (Candidate, error) {
	out := Candidate{
		Chord:    c,
		Position: pos,
		Members:  MembersLabel{Notes: c.Notes()},
		Reading:  ReadingLabel{Main: read[i].Chord.Main},
	}
	if place := relation.Places(next)[i]; place.LastInPhrase && place.FullPhrase {
		out.Cadence = &CadenceLabel{Cadence: read[i].Cadence}
	}
	if i == 0 {
		return out, nil
	}
	last, ok := next.Slots()[i-1].Content().Chord()
	if !ok {
		return out, nil
	}
	switch iv, err := relation.RootMotion(last, c); {
	case err == nil:
		out.RootMotion = &RootMotionLabel{Interval: iv}
	case !errors.Is(err, material.ErrInvalidInterval):
		return Candidate{}, fmt.Errorf("candidate: %w", err)
	}
	out.CommonTones = &CommonTonesLabel{Count: len(relation.CommonTones(last, c).PitchClasses())}
	out.Resolution = &ResolutionLabel{Resolution: read[i-1].Resolution}
	return out, nil
}
