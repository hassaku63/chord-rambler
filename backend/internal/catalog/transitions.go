package catalog

import (
	_ "embed" // the catalogue is embedded in the binary (ADR-0008, 6)
	"fmt"
	"strings"

	"github.com/hassaku63/chord-rambler/backend/internal/reading"
)

// Preference is how often the source says one function moves to another. The
// zero value is never returned by a loader.
type Preference int

// The preferences, named after what the source says.
const (
	// Common is a move the source says harmonies tend to make.
	Common Preference = iota + 1
	// LessCommon is a move the source says tends to be less common.
	LessCommon
	// Unstated is a move the source says nothing about how often it is
	// made. It is not a statement of frequency.
	Unstated
)

// String returns the name of the constant, for tests and logs.
func (p Preference) String() string {
	switch p {
	case Common:
		return "Common"
	case LessCommon:
		return "LessCommon"
	case Unstated:
		return "Unstated"
	default:
		return fmt.Sprintf("Preference(%d)", int(p))
	}
}

// Transition is one entry of the catalogue of function transitions: how often
// a chord of the function From is followed by one of the function To.
type Transition struct {
	From, To   reading.Function
	Preference Preference
	Grounds    Grounds
}

// FunctionTransitions is the catalogue of function transitions: one entry
// for every ordered pair of tonic, subdominant and dominant. Only the loaders
// make one with entries, so every value but the zero value holds all nine
// pairs.
type FunctionTransitions struct {
	source  Source
	entries [3][3]Transition // by From-1, then To-1
}

// Source returns where the catalogue comes from.
func (c FunctionTransitions) Source() Source { return c.source }

// Lookup returns the entry from one function to the next. It reports false
// when either is not tonic, subdominant or dominant, and for the zero value
// of FunctionTransitions, which holds no entries.
func (c FunctionTransitions) Lookup(from, to reading.Function) (Transition, bool) {
	i, ok := functionIndex(from)
	j, ok2 := functionIndex(to)
	if !ok || !ok2 {
		return Transition{}, false
	}
	e := c.entries[i][j]
	return e, e.Preference != 0
}

//go:embed function_transitions.json
var functionTransitionsJSON []byte

// LoadFunctionTransitions parses and checks the catalogue embedded in the
// binary.
func LoadFunctionTransitions() (FunctionTransitions, error) {
	return ParseFunctionTransitions(functionTransitionsJSON)
}

// The words of the JSON document.
var (
	functionWords = map[string]reading.Function{
		"T": reading.Tonic, "S": reading.Subdominant, "D": reading.Dominant,
	}
	preferenceWords = map[string]Preference{
		"common": Common, "less_common": LessCommon, "unstated": Unstated,
	}
	groundsWords = map[string]Grounds{"convention": Convention}
)

type transitionsDoc struct {
	Source      *sourceDoc `json:"source"`
	Transitions []struct {
		From       string `json:"from"`
		To         string `json:"to"`
		Preference string `json:"preference"`
		Grounds    string `json:"grounds"`
	} `json:"transitions"`
}

// transitionsNames are the names of the members of the document.
var transitionsNames = []string{
	"source", "url", "version", "style",
	"transitions", "from", "to", "preference", "grounds",
}

// ParseFunctionTransitions parses a catalogue of function transitions and
// checks it: the source has an address, a version and a known style, every
// value is known, and each of the nine pairs of functions appears exactly
// once. Names of members are matched exactly; a name that is not part of the
// catalogue, or one repeated in an object, is an error too. Every error wraps
// ErrInvalid, and only ErrInvalid.
func ParseFunctionTransitions(data []byte) (FunctionTransitions, error) {
	out, err := parseFunctionTransitions(data)
	if err != nil {
		return FunctionTransitions{}, fmt.Errorf("%w: function transitions: %v", ErrInvalid, err)
	}
	return out, nil
}

func parseFunctionTransitions(data []byte) (FunctionTransitions, error) {
	var doc transitionsDoc
	if err := decodeStrict(data, &doc, transitionsNames...); err != nil {
		return FunctionTransitions{}, err
	}
	source, err := parseSource(doc.Source)
	if err != nil {
		return FunctionTransitions{}, err
	}
	out := FunctionTransitions{source: source}
	var seen [3][3]bool
	for n, e := range doc.Transitions {
		t := Transition{
			From:       functionWords[e.From],
			To:         functionWords[e.To],
			Preference: preferenceWords[e.Preference],
			Grounds:    groundsWords[e.Grounds],
		}
		i, ok := functionIndex(t.From)
		j, ok2 := functionIndex(t.To)
		switch {
		case !ok || !ok2:
			return FunctionTransitions{}, fmt.Errorf("entry %d: unknown function %q → %q", n, e.From, e.To)
		case t.Preference == 0:
			return FunctionTransitions{}, fmt.Errorf("entry %d: unknown preference %q", n, e.Preference)
		case t.Grounds == 0:
			return FunctionTransitions{}, fmt.Errorf("entry %d: unknown grounds %q", n, e.Grounds)
		case seen[i][j]:
			return FunctionTransitions{}, fmt.Errorf("entry %d: %s → %s appears twice", n, e.From, e.To)
		}
		seen[i][j] = true
		out.entries[i][j] = t
	}
	if missing := missingPairs(seen); len(missing) > 0 {
		return FunctionTransitions{}, fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	return out, nil
}

// functions are tonic, subdominant and dominant, in the order of the table of
// entries.
var functions = [3]reading.Function{reading.Tonic, reading.Subdominant, reading.Dominant}

// missingPairs names the pairs of functions not seen.
func missingPairs(seen [3][3]bool) []string {
	var out []string
	for i, from := range functions {
		for j, to := range functions {
			if !seen[i][j] {
				out = append(out, fmt.Sprintf("%v → %v", from, to))
			}
		}
	}
	return out
}

// functionIndex returns the index of tonic, subdominant and dominant in the
// table of entries, and false for any other function.
func functionIndex(f reading.Function) (int, bool) {
	switch f {
	case reading.Tonic:
		return 0, true
	case reading.Subdominant:
		return 1, true
	case reading.Dominant:
		return 2, true
	case reading.NoFunction:
		return 0, false
	default:
		return 0, false
	}
}
