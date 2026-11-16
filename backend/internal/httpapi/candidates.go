package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hassaku63/chord-rambler/backend/internal/query"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// The bodies of the contract's candidate operations.
type (
	nextCandidatesBody struct {
		Progression progressionBody `json:"progression"`
		Beats       int             `json:"beats"`
	}
	// slotCandidatesBody takes the slot as a pointer, so that a body without
	// it is refused rather than read as slot 0.
	slotCandidatesBody struct {
		Progression progressionBody `json:"progression"`
		Slot        *int            `json:"slot"`
	}
)

// The labels of the contract's Candidate and SlotCandidate. The
// labels the readings also have are written by the same functions.
type (
	candidateGroup[T any] struct {
		Kind       string `json:"kind"`
		Candidates []T    `json:"candidates"`
	}
	groupsResponse[T any] struct {
		Groups []candidateGroup[T] `json:"groups"`
	}
	candidateLabels struct {
		Chord       *chordNameLabel    `json:"chord"`
		Roman       *romanLabel        `json:"roman"`
		Reading     *chordReadingLabel `json:"reading"`
		RootMotion  *rootMotionLabel   `json:"rootMotion,omitempty"`
		CommonTones *commonTonesLabel  `json:"commonTones,omitempty"`
		Resolution  *resolutionLabel   `json:"resolution,omitempty"`
		Cadence     *cadenceLabel      `json:"cadence,omitempty"`
		Preference  *preferenceLabel   `json:"preference,omitempty"`
	}
	slotCandidateLabels struct {
		candidateLabels
		NextResolution *resolutionLabel    `json:"nextResolution,omitempty"`
		NextCadence    *cadenceLabel       `json:"nextCadence,omitempty"`
		RoleFit        *candidateRoleLabel `json:"roleFit,omitempty"`
	}
	rootMotionLabel struct {
		Basis    string `json:"basis"`
		Interval string `json:"interval"`
	}
	commonTonesLabel struct {
		Basis string `json:"basis"`
		Count int    `json:"count"`
	}
	preferenceLabel struct {
		Basis      string `json:"basis"`
		From       string `json:"from"`
		To         string `json:"to"`
		Preference string `json:"preference"`
		Style      string `json:"style"`
	}
	candidateRoleLabel struct {
		Basis string   `json:"basis"`
		Fit   string   `json:"fit"`
		Unmet []string `json:"unmet"`
	}
)

// answer reads the body into B, asks the engine with ask and writes what it
// returns. A *requestError and an error of refusals are a 400 (or the status
// the request error carries); any other error is the server's (500).
func answer[B, O any](a *api, ask func(B) (O, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body B
		if err := decodeBody(w, r, &body); err != nil {
			a.writeRequestError(w, err)
			return
		}
		out, err := ask(body)
		if err != nil {
			a.writeFailure(w, err)
			return
		}
		a.writeJSON(w, http.StatusOK, "application/json", out)
	}
}

func (a *api) writeFailure(w http.ResponseWriter, err error) {
	if re := (*requestError)(nil); errors.As(err, &re) {
		a.writeRequestError(w, re)
		return
	}
	if text, ok := refusalOf(err); ok {
		a.logger.Info("the engine refused the request", "error", err)
		if se := (*slotError)(nil); errors.As(err, &se) {
			a.writeProblemIn(w, http.StatusBadRequest, fmt.Sprintf("slot %d: %s", se.slot, text), &se.slot)
			return
		}
		a.writeProblem(w, http.StatusBadRequest, text)
		return
	}
	a.logger.Error("answering the query failed", "error", err)
	a.writeProblem(w, http.StatusInternalServerError, "")
}

// slotError names the slot an engine error is about, which the engine's own
// error does not carry as a value.
type slotError struct {
	slot int
	err  error
}

func (e *slotError) Error() string { return fmt.Sprintf("slot %d: %v", e.slot, e.err) }
func (e *slotError) Unwrap() error { return e.err }

func inSlot(slot int, err error) error {
	if err == nil {
		return nil
	}
	return &slotError{slot: slot, err: err}
}

func (a *api) nextCandidates() http.HandlerFunc {
	return answer(a, func(body nextCandidatesBody) (groupsResponse[candidateLabels], error) {
		p, err := body.Progression.progression()
		if err != nil {
			return groupsResponse[candidateLabels]{}, err
		}
		groups, err := query.NextCandidates(p, body.Beats, a.transitions)
		if err != nil {
			// The slot asked about is the one after the last.
			return groupsResponse[candidateLabels]{}, inSlot(p.Len(), err)
		}
		out := groupsResponse[candidateLabels]{Groups: make([]candidateGroup[candidateLabels], 0, len(groups))}
		for _, g := range groups {
			cg, err := groupOf(g.Kind, g.Candidates, labelsOf)
			if err != nil {
				return groupsResponse[candidateLabels]{}, err
			}
			out.Groups = append(out.Groups, cg)
		}
		return out, nil
	})
}

func (a *api) slotCandidates() http.HandlerFunc {
	return answer(a, func(body slotCandidatesBody) (groupsResponse[slotCandidateLabels], error) {
		if body.Slot == nil {
			return groupsResponse[slotCandidateLabels]{}, badRequest("slot: required")
		}
		p, err := body.Progression.progression()
		if err != nil {
			return groupsResponse[slotCandidateLabels]{}, err
		}
		groups, err := query.SlotCandidates(p, *body.Slot, a.transitions)
		if err != nil {
			return groupsResponse[slotCandidateLabels]{}, inSlot(*body.Slot, err)
		}
		out := groupsResponse[slotCandidateLabels]{Groups: make([]candidateGroup[slotCandidateLabels], 0, len(groups))}
		for _, g := range groups {
			cg, err := groupOf(g.Kind, g.Candidates, slotLabelsOf)
			if err != nil {
				return groupsResponse[slotCandidateLabels]{}, err
			}
			out.Groups = append(out.Groups, cg)
		}
		return out, nil
	})
}

// groupOf writes one group of candidates with the labels of each.
func groupOf[C, O any](kind reading.Kind, cs []C, labels func(C) (O, error)) (candidateGroup[O], error) {
	out := candidateGroup[O]{Kind: kindID(kind), Candidates: make([]O, 0, len(cs))}
	for _, c := range cs {
		l, err := labels(c)
		if err != nil {
			return candidateGroup[O]{}, err
		}
		out.Candidates = append(out.Candidates, l)
	}
	return out, nil
}

// labelsOf writes the labels of a candidate. The numeral is written from the
// candidate's position.
func labelsOf(c query.Candidate) (candidateLabels, error) {
	name, err := chordNameOf(c.Chord)
	if err != nil {
		return candidateLabels{}, err
	}
	roman, err := romanAt(c.Position, c.Reading.Main)
	if err != nil {
		return candidateLabels{}, err
	}
	out := candidateLabels{
		Chord:      name,
		Roman:      roman,
		Reading:    chordReadingOf(c.Reading.Basis(), c.Reading.Main),
		Resolution: resolutionOf(c.Resolution),
		Cadence:    cadenceOf(c.Cadence),
	}
	if l := c.RootMotion; l != nil {
		out.RootMotion = &rootMotionLabel{Basis: basisID(l.Basis()), Interval: l.Interval.String()}
	}
	if l := c.CommonTones; l != nil {
		out.CommonTones = &commonTonesLabel{Basis: basisID(l.Basis()), Count: l.Count}
	}
	if l := c.Preference; l != nil {
		out.Preference = &preferenceLabel{
			Basis: basisID(l.Basis()), From: functionID(l.From), To: functionID(l.To),
			Preference: preferenceID(l.Preference), Style: styleID(l.Style),
		}
	}
	return out, nil
}

// slotLabelsOf writes the labels of a candidate for a slot to fill.
func slotLabelsOf(c query.SlotCandidate) (slotCandidateLabels, error) {
	base, err := labelsOf(c.Candidate)
	if err != nil {
		return slotCandidateLabels{}, err
	}
	out := slotCandidateLabels{
		candidateLabels: base,
		NextResolution:  resolutionOf(c.NextResolution),
		NextCadence:     cadenceOf(c.NextCadence),
	}
	if l := c.Role; l != nil {
		unmet := make([]string, 0, len(l.Unmet))
		for _, u := range l.Unmet {
			unmet = append(unmet, conditionID(u))
		}
		out.RoleFit = &candidateRoleLabel{Basis: basisID(l.Basis()), Fit: roleFitID(l.Fit), Unmet: unmet}
	}
	return out, nil
}

func resolutionOf(l *query.ResolutionLabel) *resolutionLabel {
	if l == nil {
		return nil
	}
	return &resolutionLabel{Basis: basisID(l.Basis()), Resolution: resolutionID(l.Resolution)}
}

func cadenceOf(l *query.CadenceLabel) *cadenceLabel {
	if l == nil {
		return nil
	}
	return &cadenceLabel{Basis: basisID(l.Basis()), Cadence: cadenceID(l.Cadence)}
}
