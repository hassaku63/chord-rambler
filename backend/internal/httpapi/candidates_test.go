package httpapi_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/httpapi"
)

const (
	nextCandidatesPath = "/api/v1/progressions/next-candidates"
	slotCandidatesPath = "/api/v1/progressions/slot-candidates"
)

// post sends the body to the path of a handler given the catalogue the server
// loads when it starts, and returns what was written.
func post(t *testing.T, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	transitions, err := catalog.LoadFunctionTransitions()
	if err != nil {
		t.Fatalf("loading the catalogue: %v", err)
	}
	handler := httpapi.NewHandler(httpapi.Deps{Logger: slog.New(slog.DiscardHandler), Transitions: transitions})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)
	return rec
}

// candidateGroups is the answer of both candidate routes, each candidate kept
// as JSON.
type candidateGroups struct {
	Groups []struct {
		Kind       string            `json:"kind"`
		Candidates []json.RawMessage `json:"candidates"`
	} `json:"groups"`
}

type group struct {
	kind  string
	names []string // the chord names, in order; nil when not compared
	first string   // the first candidate, as JSON
}

// checkGroups checks the status, the order of the groups, their sizes or
// names, and the first candidate of each.
func checkGroups(t *testing.T, rec *httptest.ResponseRecorder, want []group) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200\n%s", rec.Code, rec.Body)
	}
	got := decode[candidateGroups](t, rec)
	if len(got.Groups) != len(want) {
		t.Fatalf("%d groups, want %d", len(got.Groups), len(want))
	}
	for i, w := range want {
		g := got.Groups[i]
		if g.Kind != w.kind {
			t.Errorf("group %d is %q, want %q", i, g.Kind, w.kind)
			continue
		}
		if w.names != nil {
			var names []string
			for _, c := range g.Candidates {
				var n struct {
					Chord struct {
						Name string `json:"name"`
					} `json:"chord"`
				}
				if err := json.Unmarshal(c, &n); err != nil {
					t.Fatal(err)
				}
				names = append(names, n.Chord.Name)
			}
			if !slices.Equal(names, w.names) {
				t.Errorf("%s: %q, want %q", w.kind, names, w.names)
			}
		}
		if len(g.Candidates) == 0 || !sameJSON(t, string(g.Candidates[0]), w.first) {
			t.Errorf("%s: first candidate\n got: %s\nwant: %s", w.kind, g.Candidates[0], w.first)
		}
	}
}

// After G7 in C, the diatonic chords to which the dominant commonly moves (the
// tonics: I, iii, vi) come first, each by its degree, triad before seventh
// (docs/architecture/domain-model.md, 末尾の次の候補). G up to C is a fourth
// and they share G; G7 to C resolves as expected. The other groups go by the
// degree: V/ii (A) and ♭III (E♭, which phrygian, dorian and aeolian on C
// hold), each a step or a sixth from G, and G7 resolves irregularly to them.
func TestNextCandidatesAfterG7InC(t *testing.T) {
	rec := post(t, nextCandidatesPath, `{"progression": {"key": "C", "slots": [{"chord": "G7", "beats": 4}]}, "beats": 4}`)
	checkGroups(t, rec, []group{
		{"diatonic",
			[]string{"C", "Cmaj7", "Em", "Em7", "Am", "Am7", "Dm", "Dm7", "F", "Fmaj7", "G", "G7", "Bdim", "Bm7(♭5)"},
			`{"chord": {"basis": "definition", "name": "C"}, "roman": {"basis": "computation", "numeral": "I"},
			  "reading": {"basis": "school", "kind": "diatonic", "function": "tonic", "modes": [], "fromParallelMinor": false},
			  "rootMotion": {"basis": "computation", "interval": "P4"}, "commonTones": {"basis": "computation", "count": 1},
			  "resolution": {"basis": "school", "resolution": "expected"},
			  "preference": {"basis": "convention", "from": "dominant", "to": "tonic", "preference": "common", "style": "classical"}}`},
		{"secondary-dominant",
			[]string{"A", "A7", "B", "B7", "C7", "D", "D7", "E", "E7"},
			`{"chord": {"basis": "definition", "name": "A"}, "roman": {"basis": "school", "numeral": "V/ii"},
			  "reading": {"basis": "school", "kind": "secondary-dominant", "function": "none", "modes": [], "fromParallelMinor": false, "target": 2},
			  "rootMotion": {"basis": "computation", "interval": "M2"}, "commonTones": {"basis": "computation", "count": 0},
			  "resolution": {"basis": "school", "resolution": "irregular"}}`},
		{"borrowed",
			[]string{"E♭", "E♭maj7", "Fm", "Fm7", "A♭", "A♭maj7", "B♭", "B♭7"},
			`{"chord": {"basis": "definition", "name": "E♭"}, "roman": {"basis": "computation", "numeral": "♭III"},
			  "reading": {"basis": "school", "kind": "borrowed", "function": "none", "modes": ["dorian", "phrygian", "aeolian"], "fromParallelMinor": true},
			  "rootMotion": {"basis": "computation", "interval": "m6"}, "commonTones": {"basis": "computation", "count": 1},
			  "resolution": {"basis": "school", "resolution": "irregular"}}`},
	})
}

// A slot to prepare G7 in C is fitted by the subdominants, ii first; with no
// slot before it, the rest keep the order of the degrees. V/ii has no
// function, so it does not fit, and it goes on to G7 irregularly (it points
// to ii). ♭III neither, and it names no resolution.
func TestSlotCandidatesToPrepareG7InC(t *testing.T) {
	rec := post(t, slotCandidatesPath, `{"progression": {"key": "C", "slots": [
		{"content": "unchosen", "beats": 4, "role": "prepare"}, {"chord": "G7", "beats": 4}]}, "slot": 0}`)
	checkGroups(t, rec, []group{
		{"diatonic",
			[]string{"Dm", "Dm7", "F", "Fmaj7", "C", "Cmaj7", "Em", "Em7", "G", "G7", "Am", "Am7", "Bdim", "Bm7(♭5)"},
			`{"chord": {"basis": "definition", "name": "Dm"}, "roman": {"basis": "computation", "numeral": "ii"},
			  "reading": {"basis": "school", "kind": "diatonic", "function": "subdominant", "modes": [], "fromParallelMinor": false},
			  "nextResolution": {"basis": "school", "resolution": "none"},
			  "roleFit": {"basis": "school", "fit": "fits", "unmet": []}}`},
		{"secondary-dominant", nil,
			`{"chord": {"basis": "definition", "name": "A"}, "roman": {"basis": "school", "numeral": "V/ii"},
			  "reading": {"basis": "school", "kind": "secondary-dominant", "function": "none", "modes": [], "fromParallelMinor": false, "target": 2},
			  "nextResolution": {"basis": "school", "resolution": "irregular"},
			  "roleFit": {"basis": "school", "fit": "misfits", "unmet": ["function"]}}`},
		{"borrowed", nil,
			`{"chord": {"basis": "definition", "name": "E♭"}, "roman": {"basis": "computation", "numeral": "♭III"},
			  "reading": {"basis": "school", "kind": "borrowed", "function": "none", "modes": ["dorian", "phrygian", "aeolian"], "fromParallelMinor": true},
			  "nextResolution": {"basis": "school", "resolution": "none"},
			  "roleFit": {"basis": "school", "fit": "misfits", "unmet": ["function"]}}`},
	})
}

// TestCandidatesRejectWhatTheEngineRefuses checks every 400 of the candidate
// routes, one case each.
func TestCandidatesRejectWhatTheEngineRefuses(t *testing.T) {
	tests := []struct {
		name, path, body string
		wantDetail       []string
	}{
		{"no beats", nextCandidatesPath,
			`{"progression": {"key": "C", "slots": []}, "beats": 0}`, []string{"less than one beat"}},
		{"crosses a phrase", nextCandidatesPath,
			`{"progression": {"key": "C", "slots": [{"chord": "C", "beats": 12}]}, "beats": 8}`, []string{"boundary of a phrase"}},
		{"key that cannot be spelt", nextCandidatesPath,
			`{"progression": {"key": "E♯", "slots": []}, "beats": 4}`, []string{"E♯ major", "double sharp"}},
		{"progression not read", nextCandidatesPath,
			`{"progression": {"key": "C", "slots": [{"chord": "Dmx7", "beats": 4}]}, "beats": 4}`, []string{"slot 0", `"Dmx7"`}},
		{"place of a wrong type", nextCandidatesPath,
			`{"progression": {"key": "C", "slots": [{"chord": "C", "beats": "4"}]}, "beats": 4}`,
			[]string{"progression.slots[0].beats: expected integer"}},
		{"slot outside", slotCandidatesPath,
			`{"progression": {"key": "C", "slots": [{"chord": "C", "beats": 4}]}, "slot": 1}`, []string{"slot 1", "outside"}},
		{"slot left out", slotCandidatesPath,
			`{"progression": {"key": "C", "slots": [{"chord": "C", "beats": 4}]}}`, []string{"slot: required"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(t, tt.path, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400\n%s", rec.Code, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json", ct)
			}
			body := decode[struct {
				Detail string `json:"detail"`
			}](t, rec)
			for _, want := range tt.wantDetail {
				if !strings.Contains(body.Detail, want) {
					t.Errorf("detail = %q, want it to contain %q", body.Detail, want)
				}
			}
		})
	}
}
