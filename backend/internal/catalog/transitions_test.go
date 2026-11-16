package catalog_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/catalog"
	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// TestEmbeddedFunctionTransitions fixes the embedded catalogue to what its
// source says. Read in openmusictheory/openmusictheory.github.io, file
// harmonicSyntax1.md at commit a907aa015f7ec54b925ddb070d2d36aecd0dc705 (and
// the same page on the site, harmonicSyntax1.html), the third to fifth
// paragraphs before the first heading ("The idealized phrase"):
//
//   - harmonies tend to progress through the cycle T → S → D → T: Common;
//   - T to D, D to S "etc." tend to be less common, at least in classical
//     music: LessCommon. The owner reads "etc." as the other moves off the
//     cycle, so S → T is LessCommon too (#114);
//   - the page says nothing about how often a function follows itself; it
//     only speaks of zones of one function: Unstated (#114).
//
// The words "at least in classical music" are in the sentence on the less
// common moves; the catalogue keeps the limit for the whole of it, as Style
// (#114).
func TestEmbeddedFunctionTransitions(t *testing.T) {
	c, err := catalog.LoadFunctionTransitions()
	if err != nil {
		t.Fatalf("LoadFunctionTransitions() error = %v", err)
	}
	wantSource := catalog.Source{
		URL:     "https://github.com/openmusictheory/openmusictheory.github.io/blob/a907aa015f7ec54b925ddb070d2d36aecd0dc705/harmonicSyntax1.md",
		Version: "a907aa015f7ec54b925ddb070d2d36aecd0dc705",
		Style:   catalog.Classical,
	}
	if got := c.Source(); got != wantSource {
		t.Errorf("Source() = %+v; want %+v", got, wantSource)
	}
	tests := []struct {
		from, to reading.Function
		want     catalog.Preference
	}{
		{reading.Tonic, reading.Subdominant, catalog.Common},
		{reading.Subdominant, reading.Dominant, catalog.Common},
		{reading.Dominant, reading.Tonic, catalog.Common},
		{reading.Tonic, reading.Dominant, catalog.LessCommon},
		{reading.Dominant, reading.Subdominant, catalog.LessCommon},
		{reading.Subdominant, reading.Tonic, catalog.LessCommon},
		{reading.Tonic, reading.Tonic, catalog.Unstated},
		{reading.Subdominant, reading.Subdominant, catalog.Unstated},
		{reading.Dominant, reading.Dominant, catalog.Unstated},
	}
	for _, tt := range tests {
		got, ok := c.Lookup(tt.from, tt.to)
		want := catalog.Transition{From: tt.from, To: tt.to, Preference: tt.want, Grounds: catalog.Convention}
		if !ok || got != want {
			t.Errorf("Lookup(%v, %v) = %+v, %v; want %+v, true", tt.from, tt.to, got, ok, want)
		}
	}
}

func TestLookupInTheZeroValue(t *testing.T) {
	var c catalog.FunctionTransitions
	if got, ok := c.Lookup(reading.Tonic, reading.Tonic); ok {
		t.Errorf("FunctionTransitions{}.Lookup(Tonic, Tonic) = %+v, true; want false", got)
	}
}

func TestLookupOutsideTheThreeFunctions(t *testing.T) {
	c, err := catalog.LoadFunctionTransitions()
	if err != nil {
		t.Fatalf("LoadFunctionTransitions() error = %v", err)
	}
	for _, pair := range [][2]reading.Function{
		{reading.NoFunction, reading.Tonic},
		{reading.Dominant, reading.NoFunction},
		{reading.Function(9), reading.Tonic},
	} {
		if got, ok := c.Lookup(pair[0], pair[1]); ok {
			t.Errorf("Lookup(%v, %v) = %+v, true; want false", pair[0], pair[1], got)
		}
	}
}

// The nine entries of a valid catalogue, for the broken ones below.
var validEntries = []string{
	`{"from": "T", "to": "T", "preference": "unstated", "grounds": "convention"}`,
	`{"from": "T", "to": "S", "preference": "common", "grounds": "convention"}`,
	`{"from": "T", "to": "D", "preference": "less_common", "grounds": "convention"}`,
	`{"from": "S", "to": "T", "preference": "less_common", "grounds": "convention"}`,
	`{"from": "S", "to": "S", "preference": "unstated", "grounds": "convention"}`,
	`{"from": "S", "to": "D", "preference": "common", "grounds": "convention"}`,
	`{"from": "D", "to": "T", "preference": "common", "grounds": "convention"}`,
	`{"from": "D", "to": "S", "preference": "less_common", "grounds": "convention"}`,
	`{"from": "D", "to": "D", "preference": "unstated", "grounds": "convention"}`,
}

const validSource = `{"url": "https://example.org/a", "version": "v1", "style": "classical"}`

// doc writes a catalogue with the source and the entries.
func doc(source string, entries ...string) string {
	return `{"source": ` + source + `, "transitions": [` + strings.Join(entries, ", ") + `]}`
}

// replaced returns the valid entries with entry i replaced.
func replaced(i int, entry string) []string {
	out := append([]string(nil), validEntries...)
	out[i] = entry
	return out
}

func TestParseFunctionTransitions(t *testing.T) {
	if _, err := catalog.ParseFunctionTransitions([]byte(doc(validSource, validEntries...))); err != nil {
		t.Fatalf("ParseFunctionTransitions(valid) error = %v", err)
	}
	// want is part of the message the package writes, or empty when the
	// decoder of the standard library finds the error.
	tests := []struct {
		name string
		data string
		want string
	}{
		{"not JSON", `{"source": `, ""},
		{"null", `null`, "the source needs"},
		{"an array", `[]`, ""},
		{"data after the document", doc(validSource, validEntries...) + ` {}`, "data after the document"},
		{"a closing brace after the document", doc(validSource, validEntries...) + ` }`, ""},
		{"no source", `{"transitions": [` + strings.Join(validEntries, ", ") + `]}`, "the source needs"},
		{"empty url", doc(`{"url": "", "version": "v1", "style": "classical"}`, validEntries...), "the source needs"},
		{"blank version", doc(`{"url": "https://example.org/a", "version": " ", "style": "classical"}`, validEntries...), "the source needs"},
		{"no version", doc(`{"url": "https://example.org/a", "style": "classical"}`, validEntries...), "the source needs"},
		{"no style", doc(`{"url": "https://example.org/a", "version": "v1"}`, validEntries...), "unknown style"},
		{"empty style", doc(`{"url": "https://example.org/a", "version": "v1", "style": ""}`, validEntries...), "unknown style"},
		{"unknown style", doc(`{"url": "https://example.org/a", "version": "v1", "style": "pop"}`, validEntries...), "unknown style"},
		{"no entries", doc(validSource), "missing Tonic → Tonic"},
		{"a pair missing", doc(validSource, validEntries[1:]...), "missing Tonic → Tonic"},
		{"a pair twice", doc(validSource, replaced(0, validEntries[1])...), "T → S appears twice"},
		{"unknown function", doc(validSource, replaced(0, `{"from": "X", "to": "T", "preference": "unstated", "grounds": "convention"}`)...), "unknown function"},
		{"no function", doc(validSource, replaced(0, `{"to": "T", "preference": "unstated", "grounds": "convention"}`)...), "unknown function"},
		{"function in lower case", doc(validSource, replaced(0, `{"from": "t", "to": "T", "preference": "unstated", "grounds": "convention"}`)...), "unknown function"},
		{"function with a space", doc(validSource, replaced(0, `{"from": " T", "to": "T", "preference": "unstated", "grounds": "convention"}`)...), "unknown function"},
		{"unknown preference", doc(validSource, replaced(1, `{"from": "T", "to": "S", "preference": "often", "grounds": "convention"}`)...), "unknown preference"},
		{"no preference", doc(validSource, replaced(1, `{"from": "T", "to": "S", "grounds": "convention"}`)...), "unknown preference"},
		{"unknown grounds", doc(validSource, replaced(1, `{"from": "T", "to": "S", "preference": "common", "grounds": "association"}`)...), "unknown grounds"},
		{"no grounds", doc(validSource, replaced(1, `{"from": "T", "to": "S", "preference": "common"}`)...), "unknown grounds"},
		{"unknown name in an entry", doc(validSource, replaced(1, `{"from": "T", "to": "S", "preference": "common", "grounds": "convention", "weight": 1}`)...), "unknown name \"weight\""},
		{"unknown name at the top", `{"source": ` + validSource + `, "transitions": [` + strings.Join(validEntries, ", ") + `], "notes": "x"}`, "unknown name \"notes\""},
		{"unknown name in the source", doc(`{"url": "https://example.org/a", "version": "v1", "style": "classical", "license": "x"}`, validEntries...), "unknown name \"license\""},
		{"name in another case", doc(validSource, replaced(1, `{"FROM": "T", "to": "S", "preference": "common", "grounds": "convention"}`)...), "unknown name \"FROM\""},
		{"name in the wrong place", doc(validSource, replaced(1, `{"from": "T", "to": "S", "preference": "common", "grounds": "convention", "url": "x"}`)...), ""},
		{"name twice in an entry", doc(validSource, replaced(1, `{"from": "T", "to": "S", "preference": "common", "grounds": "convention", "from": "D"}`)...), "name \"from\" appears twice"},
		{"name twice at the top", `{"source": ` + validSource + `, "source": ` + validSource + `, "transitions": [` + strings.Join(validEntries, ", ") + `]}`, "name \"source\" appears twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := catalog.ParseFunctionTransitions([]byte(tt.data))
			if !errors.Is(err, catalog.ErrInvalid) || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseFunctionTransitions(%s) error = %v; want one wrapping ErrInvalid, saying %q", tt.data, err, tt.want)
			}
		})
	}
}
