package httpapi_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/httpapi"
)

const notesPath = "/api/v1/progressions/notes"

// postNotes sends the body to the notes route and returns what was written.
func postNotes(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	handler := httpapi.NewHandler(httpapi.Deps{Logger: slog.New(slog.DiscardHandler)})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, notesPath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)
	return rec
}

// TestNotesOfTwoBarsInC checks the whole answer for C, no chord and G7 over
// two bars. Each chord gives its bass and then its upper voices; the slot with
// no chord gives nothing, and G7 starts where it would anyway (beat 6).
func TestNotesOfTwoBarsInC(t *testing.T) {
	body := `{"key": "C", "slots": [
		{"chord": "C", "beats": 4},
		{"content": "none", "beats": 2},
		{"chord": "G7", "beats": 2}]}`
	// C: bass C2 = 36; upper C4 = 60, E4 = 64, G4 = 67.
	// G7: bass G2 = 43; upper G4 = 67, B4 = 71, D5 = 74, F5 = 77.
	want := `{"notes": [
		{"pitch": 36, "start": 0, "beats": 4, "velocity": 64},
		{"pitch": 60, "start": 0, "beats": 4, "velocity": 64},
		{"pitch": 64, "start": 0, "beats": 4, "velocity": 64},
		{"pitch": 67, "start": 0, "beats": 4, "velocity": 64},
		{"pitch": 43, "start": 6, "beats": 2, "velocity": 64},
		{"pitch": 67, "start": 6, "beats": 2, "velocity": 64},
		{"pitch": 71, "start": 6, "beats": 2, "velocity": 64},
		{"pitch": 74, "start": 6, "beats": 2, "velocity": 64},
		{"pitch": 77, "start": 6, "beats": 2, "velocity": 64}]}`

	rec := postNotes(t, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200\n%s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if !sameJSON(t, rec.Body.String(), want) {
		t.Errorf("notes differ\n got: %s\nwant: %s", rec.Body, want)
	}
}

// TestNotesOfNoChordIsAnEmptyList keeps notes an array, never null.
func TestNotesOfNoChordIsAnEmptyList(t *testing.T) {
	rec := postNotes(t, `{"key": "C", "slots": [{"content": "unchosen", "beats": 4}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200\n%s", rec.Code, rec.Body)
	}
	if !sameJSON(t, rec.Body.String(), `{"notes": []}`) {
		t.Errorf("body = %s, want an empty notes array", rec.Body)
	}
}

// TestNotesRejectAnUnreadableChord checks that the failure is the one of the
// readings: a 400 problem whose detail names the slot.
func TestNotesRejectAnUnreadableChord(t *testing.T) {
	rec := postNotes(t, `{"key": "C", "slots": [{"chord": "C", "beats": 4}, {"chord": "Dmx7", "beats": 4}]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400\n%s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
	var got struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("the response is not JSON: %v", err)
	}
	for _, want := range []string{"slot 1", `"Dmx7"`} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail %q lacks %s", got.Detail, want)
		}
	}
}
