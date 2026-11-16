package httpapi_test

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"
)

const variationsPath = "/api/v1/variations"

// The variations of Dm in C follow the list of a minor triad
// (docs/architecture/domain-model.md, 和音の変形); Dm itself is marked, and
// its notes are all in the key.
func TestVariationsOfDmInC(t *testing.T) {
	rec := post(t, variationsPath, `{"key": "C", "chord": "Dm"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200\n%s", rec.Code, rec.Body)
	}
	got := decode[struct {
		Variations []json.RawMessage `json:"variations"`
	}](t, rec)
	var names []string
	for _, v := range got.Variations {
		var n struct {
			Chord struct {
				Name string `json:"name"`
			} `json:"chord"`
		}
		if err := json.Unmarshal(v, &n); err != nil {
			t.Fatal(err)
		}
		names = append(names, n.Chord.Name)
	}
	if want := []string{"Dm", "Dm7", "Dm6", "Dm(add9)", "Dm7(9)", "Dsus2", "Dsus4"}; !slices.Equal(names, want) {
		t.Fatalf("variations %q, want %q", names, want)
	}
	want := `{"chord": {"basis": "definition", "name": "Dm"}, "isInput": true,
		"members": {"basis": "definition", "notes": ["D", "F", "A"]},
		"reading": {"basis": "school", "kind": "diatonic", "function": "subdominant", "modes": [], "fromParallelMinor": false},
		"tensionCautions": {"basis": "school", "cautions": []}, "outOfKey": {"basis": "computation", "notes": []}}`
	if !sameJSON(t, string(got.Variations[0]), want) {
		t.Errorf("first variation\n got: %s\nwant: %s", got.Variations[0], want)
	}
}
