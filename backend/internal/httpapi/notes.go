package httpapi

import (
	"net/http"

	"github.com/hassaku63/chord-rambler/backend/internal/performance"
)

// The Notes of the contract.
type (
	notesResponse struct {
		Notes []noteBody `json:"notes"`
	}
	noteBody struct {
		Pitch    int `json:"pitch"`
		Start    int `json:"start"`
		Beats    int `json:"beats"`
		Velocity int `json:"velocity"`
	}
)

func (a *api) notes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := decodeProgression(w, r)
		if err != nil {
			a.writeRequestError(w, err)
			return
		}
		notes := performance.Notes(p)
		out := notesResponse{Notes: make([]noteBody, len(notes))}
		for i, n := range notes {
			out.Notes[i] = noteBody{Pitch: n.Pitch, Start: n.Start, Beats: n.Beats, Velocity: n.Velocity}
		}
		a.writeJSON(w, http.StatusOK, "application/json", out)
	}
}
