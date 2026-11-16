package httpapi_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/httpapi"
)

const readingsPath = "/api/v1/progressions/readings"

// postReadings sends the body to the readings and returns what was written.
func postReadings(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	handler := httpapi.NewHandler(httpapi.Deps{Logger: slog.New(slog.DiscardHandler)})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, readingsPath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)
	return rec
}

// sameJSON reports whether two JSON documents hold the same values.
func sameJSON(t *testing.T, got, want string) bool {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("the response is not JSON: %v\n%s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("the expected value is not JSON: %v", err)
	}
	return reflect.DeepEqual(g, w)
}

// TestReadingsOfFourBarsInC checks the whole answer for progressions of four
// bars in C, and for no slots. The expected readings follow the rules of
// docs/architecture/theory-conventions.md as reading.ReadProgression,
// reading.ReadRoleFits and reading.TensionCautions document them.
func TestReadingsOfFourBarsInC(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			// ii–V–I–I: G7 resolves to C as expected. The phrase ends I after
			// I, which names no cadence.
			name: "ii-V-I-I",
			body: `{"key": "C", "slots": [
				{"chord": "Dm7", "beats": 4, "role": "prepare"},
				{"chord": "G7", "beats": 4, "role": "rise"},
				{"chord": "C", "beats": 4, "role": "resolve"},
				{"chord": "C", "beats": 4, "role": "close"}]}`,
			want: `{"slots": [
				{"chord": {"basis": "definition", "name": "Dm7"}, "roman": {"basis": "computation", "numeral": "ii7"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "subdominant", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}, "roleFit": {"basis": "school", "fit": "fits"}},
				{"chord": {"basis": "definition", "name": "G7"}, "roman": {"basis": "computation", "numeral": "V7"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "dominant", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "expected"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}, "roleFit": {"basis": "school", "fit": "fits"}},
				{"chord": {"basis": "definition", "name": "C"}, "roman": {"basis": "computation", "numeral": "I"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "tonic", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}, "roleFit": {"basis": "school", "fit": "fits"}},
				{"chord": {"basis": "definition", "name": "C"}, "roman": {"basis": "computation", "numeral": "I"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "tonic", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}, "roleFit": {"basis": "school", "fit": "misfits"}}]}`,
		},
		{
			// A secondary dominant (D7, written V7/V on the school), a borrowed
			// chord from the parallel minor (Fm) and a chord with a bass (G/B,
			// whose numeral leaves the bass out) ending on a half cadence.
			name: "borrowed chord",
			body: `{"key": "C", "slots": [
				{"chord": "C", "beats": 4},
				{"chord": "D7", "beats": 4},
				{"chord": "Fm", "beats": 4},
				{"chord": "G/B", "beats": 4}]}`,
			want: `{"slots": [
				{"chord": {"basis": "definition", "name": "C"}, "roman": {"basis": "computation", "numeral": "I"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "tonic", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}},
				{"chord": {"basis": "definition", "name": "D7"}, "roman": {"basis": "school", "numeral": "V7/V"},
				 "reading": {"basis": "school", "kind": "secondary-dominant", "function": "none", "modes": [], "fromParallelMinor": false, "target": 5},
				 "resolution": {"basis": "school", "resolution": "irregular"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}},
				{"chord": {"basis": "definition", "name": "Fm"}, "roman": {"basis": "computation", "numeral": "iv"},
				 "reading": {"basis": "school", "kind": "borrowed", "function": "none", "modes": ["phrygian", "aeolian", "locrian"], "fromParallelMinor": true},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}},
				{"chord": {"basis": "definition", "name": "G/B"}, "roman": {"basis": "computation", "numeral": "V"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "dominant", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "half"}, "tensionCautions": {"basis": "school", "cautions": []}}]}`,
		},
		{
			// Slots with no chord: one that sounds nothing, and one not chosen
			// yet whose role cannot be judged. A colour role gets no role fit.
			// The 11 of G7(11) is a half step above its third.
			name: "no chord",
			body: `{"key": "C", "slots": [
				{"chord": "C", "beats": 4, "role": "stable"},
				{"content": "none", "beats": 4},
				{"content": "unchosen", "beats": 4, "role": "resolve"},
				{"chord": "G7(11)", "beats": 4, "role": "float"}]}`,
			want: `{"slots": [
				{"chord": {"basis": "definition", "name": "C"}, "roman": {"basis": "computation", "numeral": "I"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "tonic", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "none"}, "tensionCautions": {"basis": "school", "cautions": []}, "roleFit": {"basis": "school", "fit": "fits"}},
				{},
				{"roleFit": {"basis": "school", "fit": "undecided"}},
				{"chord": {"basis": "definition", "name": "G7(11)"}, "roman": {"basis": "computation", "numeral": "V7(11)"},
				 "reading": {"basis": "school", "kind": "diatonic", "function": "dominant", "modes": [], "fromParallelMinor": false},
				 "resolution": {"basis": "school", "resolution": "none"}, "cadence": {"basis": "school", "cadence": "half"}, "tensionCautions": {"basis": "school", "cautions": [{"tension": "P11", "chordTone": "M3"}]}}]}`,
		},
		{name: "no slots", body: `{"key": "F♯", "slots": []}`, want: `{"slots": []}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postReadings(t, tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d\n%s", rec.Code, http.StatusOK, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}
			if got := rec.Body.String(); !sameJSON(t, got, tt.want) {
				t.Errorf("body =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

// TestReadingsRejectWhatIsNotAProgression checks every 400 of the contract.
// The detail names the slot, counted from 0, and what could not be read.
func TestReadingsRejectWhatIsNotAProgression(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantDetail []string
	}{
		{"unreadable chord name", `{"key": "C", "slots": [{"chord": "C", "beats": 4}, {"chord": "Dmx7", "beats": 4}]}`,
			[]string{"slot 1", `"mx7"`, `"Dmx7"`}},
		{"unknown role", `{"key": "C", "slots": [{"chord": "C", "beats": 4, "role": "tonic"}]}`,
			[]string{"slot 0", `"tonic"`}},
		{"chord and content", `{"key": "C", "slots": [{"chord": "C", "content": "none", "beats": 4}]}`,
			[]string{"slot 0", "chord", "content"}},
		{"neither chord nor content", `{"key": "C", "slots": [{"beats": 4}]}`,
			[]string{"slot 0", "chord", "content"}},
		{"unknown content", `{"key": "C", "slots": [{"content": "rest", "beats": 4}]}`,
			[]string{"slot 0", `"rest"`}},
		{"no beats", `{"key": "C", "slots": [{"chord": "C", "beats": 0}]}`,
			[]string{"slot 0", "beat"}},
		{"crosses a phrase", `{"key": "C", "slots": [{"chord": "C", "beats": 12}, {"chord": "G", "beats": 8}]}`,
			[]string{"slot 1", "phrase"}},
		{"unreadable key", `{"key": "H", "slots": []}`,
			[]string{"key"}},
		{"key with no major scale", `{"key": "B𝄪", "slots": []}`,
			[]string{"key"}},
		{"key left out", `{"slots": []}`,
			[]string{"key"}},
		{"not JSON", `{"key": "C", "slots": [`,
			[]string{"JSON"}},
		{"more after the JSON", `{"key": "C", "slots": []} x`,
			[]string{"JSON"}},
		{"beats not an integer", `{"key": "C", "slots": [{"chord": "C", "beats": 2.5}]}`,
			[]string{"JSON"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postReadings(t, tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d\n%s", rec.Code, http.StatusBadRequest, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/problem+json")
			}
			body := decode[struct {
				Title  string `json:"title"`
				Status int    `json:"status"`
				Detail string `json:"detail"`
			}](t, rec)
			if body.Status != http.StatusBadRequest || body.Title != http.StatusText(http.StatusBadRequest) {
				t.Errorf("status, title = %d, %q, want %d, %q",
					body.Status, body.Title, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			}
			for _, want := range tt.wantDetail {
				if !strings.Contains(body.Detail, want) {
					t.Errorf("detail = %q, want it to contain %q", body.Detail, want)
				}
			}
		})
	}
}

// TestReadingsWriteTheRomanNumeralAsRead checks the numeral of one chord in C:
// a secondary dominant is V/x whatever its root (#83), a bass is not written
// (docs/architecture/theory-conventions.md, 6), and a root two half steps
// from its degree has no numeral.
func TestReadingsWriteTheRomanNumeralAsRead(t *testing.T) {
	tests := []struct {
		chord     string
		want      string // "" for no numeral
		wantBasis string
	}{
		{"C7", "V7/IV", "school"},
		{"A7/C♯", "V7/ii", "school"},
		{"D7(9)", "V7(9)/V", "school"},
		{"G/B", "V", "computation"},
		{"B♭", "♭VII", "computation"},
		{"Bdim7", "vii°7", "computation"},
		{"Csus4/G", "Isus4", "computation"},
		{"D𝄫", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.chord, func(t *testing.T) {
			rec := postReadings(t, `{"key": "C", "slots": [{"chord": "`+tt.chord+`", "beats": 4}]}`)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d\n%s", rec.Code, http.StatusOK, rec.Body)
			}
			body := decode[struct {
				Slots []struct {
					Roman *struct {
						Basis   string `json:"basis"`
						Numeral string `json:"numeral"`
					} `json:"roman"`
				} `json:"slots"`
			}](t, rec)
			roman := body.Slots[0].Roman
			switch {
			case tt.want == "" && roman != nil:
				t.Errorf("roman = %+v, want none", *roman)
			case tt.want == "":
			case roman == nil:
				t.Errorf("roman = none, want %q", tt.want)
			case roman.Numeral != tt.want || roman.Basis != tt.wantBasis:
				t.Errorf("roman = %q (%s), want %q (%s)", roman.Numeral, roman.Basis, tt.want, tt.wantBasis)
			}
		})
	}
}
