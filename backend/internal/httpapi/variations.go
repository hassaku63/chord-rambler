package httpapi

import (
	"net/http"

	"github.com/hassaku63/chord-rambler/backend/internal/material"
	"github.com/hassaku63/chord-rambler/backend/internal/notation"
	"github.com/hassaku63/chord-rambler/backend/internal/query"
)

// The body and the labels of the contract's variation operation.
type (
	variationsBody struct {
		Key   string `json:"key"`
		Chord string `json:"chord"`
	}
	variationsResponse struct {
		Variations []variationLabels `json:"variations"`
	}
	variationLabels struct {
		Chord           *chordNameLabel       `json:"chord"`
		IsInput         bool                  `json:"isInput"`
		Members         notesLabel            `json:"members"`
		Reading         *chordReadingLabel    `json:"reading"`
		TensionCautions *tensionCautionsLabel `json:"tensionCautions"`
		OutOfKey        notesLabel            `json:"outOfKey"`
	}
	notesLabel struct {
		Basis string   `json:"basis"`
		Notes []string `json:"notes"`
	}
)

func (a *api) variations() http.HandlerFunc {
	return answer(a, func(body variationsBody) (variationsResponse, error) {
		key, err := parseKey(body.Key)
		if err != nil {
			return variationsResponse{}, err
		}
		c, err := notation.ParseChord(body.Chord)
		if err != nil {
			return variationsResponse{}, badRequest("chord: %v", err)
		}
		vs, err := query.Variations(c, key)
		if err != nil {
			return variationsResponse{}, err
		}
		out := variationsResponse{Variations: make([]variationLabels, 0, len(vs))}
		for _, v := range vs {
			name, err := chordNameOf(v.Chord)
			if err != nil {
				return variationsResponse{}, err
			}
			out.Variations = append(out.Variations, variationLabels{
				Chord:           name,
				IsInput:         v.IsInput,
				Members:         notesOf(v.Members.Basis(), v.Members.Notes),
				Reading:         chordReadingOf(v.Reading.Basis(), v.Reading.Main),
				TensionCautions: tensionCautionsOf(v.Cautions.Basis(), v.Cautions.Cautions),
				OutOfKey:        notesOf(v.OutOfKey.Basis(), v.OutOfKey.Notes),
			})
		}
		return out, nil
	})
}

// notesOf writes note names, always as a list.
func notesOf(basis query.Basis, notes []material.NoteName) notesLabel {
	out := notesLabel{Basis: basisID(basis), Notes: make([]string, 0, len(notes))}
	for _, n := range notes {
		out.Notes = append(out.Notes, n.String())
	}
	return out
}
